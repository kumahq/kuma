package meshaccesslog

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
)

func TestPlugin() {
	meshName := "meshaccesslog"
	externalServiceName := "meshaccesslog-" + externalservice.TcpSink
	externalServiceDeployment := "externalservice-" + externalServiceName

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "echo-v1"}))).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Setup(env.Cluster)).To(Succeed())
	})
	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	// Always have new MeshAccessLog resources and log sink
	BeforeEach(func() {
		Expect(NewClusterSetup().
			Install(
				externalservice.Install(externalServiceName, externalservice.UniversalTCPSink),
			).Setup(env.Cluster),
		).To(Succeed())
	})
	E2EAfterEach(func() {
		items, err := env.Cluster.GetKumactlOptions().KumactlList("meshaccesslogs", meshName)
		Expect(err).ToNot(HaveOccurred())

		for _, item := range items {
			err := env.Cluster.GetKumactlOptions().KumactlDelete("meshaccesslog", item, meshName)
			Expect(err).ToNot(HaveOccurred())
		}

		sinkDeployment := env.Cluster.Deployment(externalServiceDeployment).(*externalservice.UniversalDeployment)
		Expect(sinkDeployment.Delete(env.Cluster)).To(Succeed())
	})

	trafficLogFormat := "%START_TIME(%s)%,%KUMA_SOURCE_SERVICE%,%KUMA_DESTINATION_SERVICE%"
	expectTrafficLogged := func() (src, dst string) {
		sinkDeployment := env.Cluster.Deployment(externalServiceDeployment).(*externalservice.UniversalDeployment)

		Eventually(func(g Gomega) {
			_, _, err := env.Cluster.ExecWithRetries("", "", AppModeDemoClient,
				"curl", "-v", "--fail", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())

			stdout, _, err := sinkDeployment.Exec("", "", "head", "-1", "/nc.out")
			g.Expect(err).ToNot(HaveOccurred())
			parts := strings.Split(stdout, ",")
			g.Expect(parts).To(HaveLen(3))

			startTimeInt, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			Expect(err).ToNot(HaveOccurred())
			startTime := time.Unix(int64(startTimeInt), 0)
			Expect(startTime).To(BeTemporally("~", time.Now(), time.Hour))

			src, dst = parts[1], parts[2]
		}, "30s", "1ms").Should(Succeed())

		return strings.TrimSpace(src), strings.TrimSpace(dst)
	}

	It("should log outgoing traffic", func() {
		yaml := fmt.Sprintf(`
type: MeshAccessLog
name: client-outgoing
mesh: meshaccesslog
spec:
 targetRef:
   kind: MeshService
   name: demo-client
 to:
   - targetRef:
       kind: Mesh
     default:
       backends:
       - tcp:
           format:
             plain: '%s'
           address: "%s_%s:9999"
`, trafficLogFormat, env.Cluster.Name(), externalServiceDeployment)
		Expect(YamlUniversal(yaml)(env.Cluster)).To(Succeed())

		src, dst := expectTrafficLogged()

		Expect(src).To(Equal(AppModeDemoClient))
		Expect(dst).To(Equal("test-server"))
	})

	It("should log incoming traffic", func() {
		yaml := fmt.Sprintf(`
type: MeshAccessLog
name: server-outgoing
mesh: meshaccesslog
spec:
 targetRef:
   kind: MeshService
   name: test-server
 from:
   - targetRef:
       kind: Mesh
     default:
       backends:
       - tcp:
           format:
             plain: '%s'
           address: "%s_%s:9999"
`, trafficLogFormat, env.Cluster.Name(), externalServiceDeployment)

		Expect(YamlUniversal(yaml)(env.Cluster)).To(Succeed())

		src, dst := expectTrafficLogged()

		Expect(src).To(Equal("unknown"))
		Expect(dst).To(Equal("test-server"))
	})
}
