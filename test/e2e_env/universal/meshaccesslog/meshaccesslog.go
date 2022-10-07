package meshaccesslog

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
)

func TestPlugin() {
	meshName := "meshaccesslog"
	externalServiceName := "meshaccesslog-" + externalservice.TcpSink
	externalServiceDeployment := "externalservice-" + externalServiceName
	var externalServiceDockerName string

	BeforeAll(func() {
		externalServiceDockerName = fmt.Sprintf("%s_%s-%s", env.Cluster.Name(), meshName, "test-server")
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(TestServerUniversal(
				"test-server", meshName, WithArgs([]string{"echo", "--instance", "echo-v1"}), WithDockerContainerName(externalServiceDockerName)),
			).
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
	expectTrafficLogged := func(target string, curlErr gomega_types.GomegaMatcher) (src, dst string) {
		sinkDeployment := env.Cluster.Deployment(externalServiceDeployment).(*externalservice.UniversalDeployment)

		Eventually(func(g Gomega) {
			_, _, err := env.Cluster.Exec("", "", AppModeDemoClient,
				"curl", "-v", "--fail", target)
			g.Expect(err).To(curlErr)

			stdout, _, err := sinkDeployment.Exec("", "", "head", "-1", "/nc.out")
			g.Expect(err).ToNot(HaveOccurred())
			parts := strings.Split(stdout, ",")
			g.Expect(parts).To(HaveLen(3))

			startTimeInt, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			Expect(err).ToNot(HaveOccurred())
			startTime := time.Unix(int64(startTimeInt), 0)
			Expect(startTime).To(BeTemporally("~", time.Now(), time.Hour))

			src, dst = parts[1], parts[2]
		}, "30s", "1s").Should(Succeed())

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

		src, dst := expectTrafficLogged("test-server.mesh", Succeed())

		Expect(src).To(Equal(AppModeDemoClient))
		Expect(dst).To(Equal("test-server"))
	})

	// This is flaky, may have something to do with access-log-streamer
	It("should log outgoing passthrough traffic", FlakeAttempts(3), func() {
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

		// 52 is empty response but the TCP connection succeeded
		src, dst := expectTrafficLogged(externalServiceDockerName, ContainSubstring("exit status 52"))

		Expect(src).To(Equal(AppModeDemoClient))
		Expect(dst).To(Equal("external"))
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

		src, dst := expectTrafficLogged("test-server.mesh", Succeed())

		Expect(src).To(Equal("unknown"))
		Expect(dst).To(Equal("test-server"))
	})
}
