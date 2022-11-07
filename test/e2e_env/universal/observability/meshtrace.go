package observability

import (
	"fmt"
	"net"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	"github.com/kumahq/kuma/test/e2e_env/universal/gateway"
	. "github.com/kumahq/kuma/test/framework"
	obs "github.com/kumahq/kuma/test/framework/deployments/observability"
)

func traceAll(meshName string, url string) string {
	return fmt.Sprintf(`
type: MeshTrace
name: trace-all
mesh: %s
spec:
  targetRef:
    kind: Mesh
  default:
    tags:
      - name: team
        literal: core
    backends:
      - zipkin:
          url: %s
`, meshName, url)
}

func PluginTest() {
	mesh := "meshtrace"
	obsDeployment := "obs-meshtrace"
	var obsClient obs.Observability

	GatewayAddressPort := func(appName string, port int) string {
		ip := env.Cluster.GetApp(appName).GetIP()
		return net.JoinHostPort(ip, strconv.Itoa(port))
	}

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(obs.Install(obsDeployment)).
			Install(MeshUniversal(mesh)).
			Install(TestServerUniversal("test-server", mesh, WithArgs([]string{"echo", "--instance", "universal1"}))).
			Install(DemoClientUniversal(AppModeDemoClient, mesh, WithTransparentProxy(true))).
			Install(GatewayProxyUniversal(mesh, "edge-gateway")).
			Install(YamlUniversal(gateway.MkGateway("edge-gateway", mesh, false, "example.kuma.io", "test-server", 8080))).
			Install(gateway.GatewayClientAppUniversal("gateway-client")).
			Setup(env.Cluster)
		obsClient = obs.From(obsDeployment, env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh)).To(Succeed())
		Expect(env.Cluster.DeleteDeployment(obsDeployment)).To(Succeed())
	})

	It("should emit traces to jaeger", func() {
		// given MeshTrace and with tracing backend
		err := YamlUniversal(traceAll(mesh, obsClient.ZipkinCollectorURL()))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() ([]string, error) {
			// when client sends requests to server
			_, _, err := env.Cluster.Exec("", "", "demo-client", "curl", "-v", "-m", "3", "--fail", "test-server.mesh")
			if err != nil {
				return nil, err
			}
			// then traces are published
			return obsClient.TracedServices()
		}, "30s", "1s").Should(ContainElements([]string{
			"demo-client",
			"jaeger-query",
			"test-server",
		}))
	})

	It("should emit MeshGateway traces to jaeger", func() {
		err := YamlUniversal(traceAll(mesh, obsClient.ZipkinCollectorURL()))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() ([]string, error) {
			gateway.ProxySimpleRequests(env.Cluster, "universal1",
				GatewayAddressPort("edge-gateway", 8080), "example.kuma.io")
			return obsClient.TracedServices()
		}, "30s", "1s").Should(ContainElements([]string{
			"edge-gateway",
			"jaeger-query",
			"test-server",
		}))
	})
}
