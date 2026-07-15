package observability

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	obs "github.com/kumahq/kuma/v3/test/framework/deployments/observability"
	"github.com/kumahq/kuma/v3/test/framework/envs/universal"
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
      - type: Zipkin
        zipkin:
          url: %s
`, meshName, url)
}

func PluginTest() {
	mesh := "meshtrace"
	obsDeployment := "obs-meshtrace"
	var obsClient obs.Observability

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(obs.Install(obsDeployment)).
			Install(MeshUniversal(mesh)).
			Install(TestServerUniversal("test-server", mesh, WithArgs([]string{"echo", "--instance", "universal1"}))).
			Install(DemoClientUniversal(AppModeDemoClient, mesh, WithTransparentProxy(true))).
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
		// given MeshTrace and with tracing backend
		err := YamlUniversal(traceAll(mesh, obsClient.ZipkinCollectorURL()))(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() ([]string, error) {
			// when client sends requests to server
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.svc.mesh.local",
			)
			if err != nil {
				return nil, err
			}
			// then traces are published
			return obsClient.TracedServices()
		}, "30s", "1s").Should(ContainElements([]string{
			"demo-client",
			"jaeger-all-in-one",
			"test-server",
		}))
	})
}
