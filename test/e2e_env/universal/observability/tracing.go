package observability

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	obs "github.com/kumahq/kuma/test/framework/deployments/observability"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func meshWithTracing(meshName, zipkinURL string) string {
	return fmt.Sprintf(`
type: Mesh
name: %s
tracing:
  defaultBackend: zipkin
  backends:
  - name: zipkin
    type: zipkin
    conf:
      url: %s
`, meshName, zipkinURL)
}

func traceAllUniversal(meshName string) string {
	return fmt.Sprintf(`
type: TrafficTrace
name: traffic-trace-all
mesh: %s
selectors:
- match:
   kuma.io/service: "*"
`, meshName)
}

func Tracing() {
	mesh := "tracing"
	obsDeployment := "obs-tracing"
	var obsClient obs.Observability

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(obs.Install(obsDeployment)).
			Install(MeshUniversal(mesh)).
			Install(TestServerUniversal("test-server", mesh, WithArgs([]string{"echo", "--instance", "universal1"}))).
			Install(DemoClientUniversal(AppModeDemoClient, mesh, WithTransparentProxy(true))).
			Install(TrafficRouteUniversal(mesh)).
			Install(TrafficPermissionUniversal(mesh)).
			Setup(universal.Cluster)
		obsClient = obs.From(obsDeployment, universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, mesh)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteDeployment(obsDeployment)).To(Succeed())
	})

	It("should emit traces to jaeger", func() {
		// given TrafficTrace and mesh with tracing backend
		err := YamlUniversal(meshWithTracing(mesh, obsClient.ZipkinCollectorURL()))(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
		err = YamlUniversal(traceAllUniversal(mesh))(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() ([]string, error) {
			// when client sends requests to server
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			if err != nil {
				return nil, err
			}
			// then traces are published
			return obsClient.TracedServices()
		}, "30s", "1s").Should(Equal([]string{
			"demo-client",
			"jaeger-all-in-one",
			"test-server",
		}))
	})
}
