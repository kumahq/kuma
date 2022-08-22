package trafficlog

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
		validLoggingBackend := fmt.Sprintf(loggingBackend, meshName, "kuma-3_externalservice-tcp-sink:9999")
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
			err := NewClusterSetup().
				Install(YamlUniversal(invalidLoggingBackend)). // start with invalid backend to test that it does not break anything
				Install(YamlUniversal(trafficLog)).
				Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "universal-1"}))).
				Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
				Install(externalservice.Install(externalservice.TcpSink, externalservice.UniversalTCPSink)).
				Setup(env.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		E2EAfterAll(func() {
			Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
			Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		It("should send a traffic log to TCP port", func() {
			// given traffic between apps with invalid logging backend
			Expect(env.Cluster.Install(YamlUniversal(invalidLoggingBackend))).To(Succeed())
			_, _, err := env.Cluster.ExecWithRetries("", "", AppModeDemoClient,
				"curl", "-v", "--fail", "test-server.mesh")
			Expect(err).ToNot(HaveOccurred())

			// when valid backend is present
			Expect(env.Cluster.Install(YamlUniversal(validLoggingBackend))).To(Succeed())

			// and traffic is sent between applications
			var startTimeStr, src, dst string
			sinkDeployment := env.Cluster.Deployment("externalservice-tcp-sink").(*externalservice.UniversalDeployment)
			Eventually(func(g Gomega) {
				_, _, err := env.Cluster.ExecWithRetries("", "", AppModeDemoClient,
					"curl", "-v", "--fail", "test-server.mesh")
				g.Expect(err).ToNot(HaveOccurred())

				stdout, _, err := sinkDeployment.Exec("", "", "head", "-1", "/nc.out")
				g.Expect(err).ToNot(HaveOccurred())
				parts := strings.Split(stdout, ",")
				g.Expect(parts).To(HaveLen(3))
				startTimeStr, src, dst = parts[0], parts[1], parts[2]
			}, "30s", "1ms").Should(Succeed())

			// then
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
