package trafficlog

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
)

func TCPLogging() {
	Describe("TrafficLog Logging to TCP", func() {
		meshName := "trafficlog-tcp-logging"
		loggingBackend := fmt.Sprintf(`
type: Mesh
name: %s
`, meshName) + `logging:
  defaultBackend: netcat
  backends:
  - name: netcat
    format: '%START_TIME(%s)%,%KUMA_SOURCE_SERVICE%,%KUMA_DESTINATION_SERVICE%'
    type: tcp
    conf:
      address: kuma-3_externalservice-tcp-sink:9999
`

		trafficLog := fmt.Sprintf(`
type: TrafficLog
name: all-traffic
mesh: %s
sources:
- match:
    kuma.io/service: "*"
destinations:
 - match:
    kuma.io/service: "*"
`, meshName)
		BeforeAll(func() {
			err := NewClusterSetup().
				Install(MeshUniversal(meshName)).
				Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "universal-1"}))).
				Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
				Install(externalservice.Install(externalservice.TcpSink, externalservice.UniversalTCPSink)).
				Install(YamlUniversal(loggingBackend)).
				Install(YamlUniversal(trafficLog)).
				Setup(env.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		E2EAfterAll(func() {
			Expect(env.Cluster.DeleteDeployment("externalservice-tcp-sink")).To(Succeed())
			Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
			Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		It("should send a traffic log to TCP port", func() {
			var startTimeStr, src, dst string
			var err error
			var stdout string
			sinkDeployment := env.Cluster.Deployment("externalservice-tcp-sink").(*externalservice.UniversalDeployment)
			Eventually(func() error {
				stdout, _, err = env.Cluster.ExecWithRetries("", "", AppModeDemoClient,
					"curl", "-v", "--fail", "test-server.mesh")
				if err != nil {
					return err
				}
				stdout, _, err = sinkDeployment.Exec("", "", "head", "-1", "/nc.out")
				if err != nil {
					return err
				}
				parts := strings.Split(stdout, ",")
				if len(parts) != 3 {
					return errors.Errorf("unexpected number of fields in: %s", stdout)
				}
				startTimeStr, src, dst = parts[0], parts[1], parts[2]
				return nil
			}, "30s", "1ms").ShouldNot(HaveOccurred())
			Expect(err).ToNot(HaveOccurred())
			startTimeInt, err := strconv.Atoi(startTimeStr)
			Expect(err).ToNot(HaveOccurred())
			startTime := time.Unix(int64(startTimeInt), 0)

			// Just testing that it is a timestamp, not accuracy. If it's
			// an int that would represent Unix time within an hour of now
			// it's probably a timestamp substitution.
			Expect(startTime).To(BeTemporally("~", time.Now(), time.Hour))

			Expect(src).To(Equal(AppModeDemoClient))
			Expect(strings.TrimSpace(dst)).To(Equal("test-server"))
		})
	})
}
