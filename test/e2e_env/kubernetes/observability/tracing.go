package observability

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	obs "github.com/kumahq/kuma/test/framework/deployments/observability"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func meshWithTracing(name, zipkinURL string) string {
	return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
spec:
  tracing:
    defaultBackend: zipkin
    backends:
    - name: zipkin
      type: zipkin
      conf:
        url: %s
`, name, zipkinURL)
}

func trafficTrace(mesh, namespace string) string {
	return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: TrafficTrace
mesh: %s
metadata:
  namespace: %s
  name: trace-all
spec:
  selectors:
  - match:
      kuma.io/service: '*'
`, mesh, namespace)
}

func Tracing() {
	ns := "tracing"
	obsNs := "obs-tracing"
	obsDeployment := "obs-tracing-deployment"
	mesh := "tracing"

	var obsClient obs.Observability
	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(ns)).
			Install(MeshKubernetes(mesh)).
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(ns), democlient.WithMesh(mesh)),
				testserver.Install(testserver.WithMesh(mesh), testserver.WithNamespace(ns)),
				obs.Install(obsDeployment, obs.WithNamespace(obsNs), obs.WithComponents(obs.JaegerComponent)),
			)).
			Install(TrafficRouteKubernetes(mesh)).
			Install(TrafficPermissionKubernetes(mesh)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
		obsClient = obs.From(obsDeployment, kubernetes.Cluster)
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, mesh, ns, obsNs)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(ns)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteDeployment(obsDeployment)).To(Succeed())
	})

	It("should emit traces to jaeger", func() {
		// given TrafficTrace and mesh with tracing backend
		err := YamlK8s(meshWithTracing(mesh, obsClient.ZipkinCollectorURL()))(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
		err = YamlK8s(trafficTrace(mesh, ns))(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", "test-server",
				client.FromKubernetesPod(ns, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			// then traces are published
			srvs, err := obsClient.TracedServices()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(srvs).To(Equal([]string{
				fmt.Sprintf("demo-client_%s_svc", ns),
				"jaeger-query",
				fmt.Sprintf("test-server_%s_svc_80", ns),
			}))
		}, "30s", "1s").Should(Succeed())
	})
}
