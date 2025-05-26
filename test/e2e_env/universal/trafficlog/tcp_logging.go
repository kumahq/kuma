package trafficlog

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func TCPLogging() {
	Describe("TrafficLog Logging to TCP", func() {
		meshName := "trafficlog-tcp-logging"
		loggingBackend := `
type: Mesh
name: %s
logging:
  defaultBackend: netcat
  backends:
  - name: netcat
    format: '%%START_TIME(%%s)%%,%%KUMA_SOURCE_SERVICE%%,%%KUMA_DESTINATION_SERVICE%%'
    type: tcp
    conf:
      address: %s
`
		validLoggingBackend := fmt.Sprintf(loggingBackend, meshName, "kuma-3_trafficlog-tcp-logging_tcp-sink:9999")
		invalidLoggingBackend := fmt.Sprintf(loggingBackend, meshName, "127.0.0.1:20")

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
			tcpSinkDockerName := fmt.Sprintf("%s_%s_%s", universal.Cluster.Name(), meshName, AppModeTcpSink)
			err := NewClusterSetup().
				Install(YamlUniversal(invalidLoggingBackend)). // start with invalid backend to test that it does not break anything
				Install(YamlUniversal(trafficLog)).
				Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "universal-1"}))).
				Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
				Install(TcpSinkUniversal(AppModeTcpSink, WithDockerContainerName(tcpSinkDockerName))).
				Install(TimeoutUniversal(meshName)).
				Install(RetryUniversal(meshName)).
				Install(TrafficRouteUniversal(meshName)).
				Install(TrafficPermissionUniversal(meshName)).
				Install(CircuitBreakerUniversal(meshName)).
				Setup(universal.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEachFailure(func() {
			DebugUniversal(universal.Cluster, meshName)
		})

		E2EAfterAll(func() {
			Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
			Expect(universal.Cluster.DeleteApp(AppModeTcpSink)).To(Succeed())
			Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		It("should send a traffic log to TCP port", func() {
			// given traffic between apps with invalid logging backend
			Expect(universal.Cluster.Install(YamlUniversal(invalidLoggingBackend))).To(Succeed())
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					universal.Cluster, AppModeDemoClient, "test-server.mesh",
				)
				g.Expect(err).ToNot(HaveOccurred())
			}).Should(Succeed())

			// when valid backend is present
			Expect(universal.Cluster.Install(YamlUniversal(validLoggingBackend))).To(Succeed())

			// and traffic is sent between applications
			var startTimeStr, src, dst string
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					universal.Cluster, AppModeDemoClient, "test-server.mesh",
				)
				g.Expect(err).ToNot(HaveOccurred())

				stdout, _, err := universal.Cluster.Exec("", "", AppModeTcpSink, "tail", "-1", "/nc.out")
				g.Expect(err).ToNot(HaveOccurred())
				parts := strings.Split(stdout, ",")
				g.Expect(parts).To(HaveLen(3))
				startTimeStr, src, dst = parts[0], parts[1], parts[2]
			}, "30s", "1s").Should(Succeed())

			// then
			startTimeInt, err := strconv.Atoi(startTimeStr)
			Expect(err).ToNot(HaveOccurred())
			startTime := time.Unix(int64(startTimeInt), 0)

			Expect(startTime).To(BeTemporally("~", time.Now(), time.Minute))

			Expect(src).To(Equal(AppModeDemoClient))
			Expect(strings.TrimSpace(dst)).To(Equal("test-server"))
		})
	})
}
