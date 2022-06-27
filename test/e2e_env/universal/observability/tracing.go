package observability

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
	obs "github.com/kumahq/kuma/test/framework/deployments/observability"
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

func traceAll(meshName string) string {
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
		// given TrafficTrace and mesh with tracing backend
		err := YamlUniversal(meshWithTracing(mesh, obsClient.ZipkinCollectorURL()))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
		err = YamlUniversal(traceAll(mesh))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() ([]string, error) {
			// when client sends requests to server
			_, _, err := env.Cluster.Exec("", "", "demo-client", "curl", "-v", "-m", "3", "--fail", "test-server.mesh")
			if err != nil {
				return nil, err
			}
			// then traces are published
			return obsClient.TracedServices()
		}, "30s", "1s").Should(Equal([]string{
			"demo-client",
			"jaeger-query",
			"test-server",
		}))
	})
}
