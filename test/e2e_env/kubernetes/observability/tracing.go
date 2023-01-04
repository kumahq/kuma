package observability

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	obs "github.com/kumahq/kuma/test/framework/deployments/observability"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
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
			Install(DemoClientK8s(mesh, ns)).
			Install(testserver.Install(testserver.WithMesh(mesh), testserver.WithNamespace(ns))).
			Install(obs.Install(obsDeployment, obs.WithNamespace(obsNs), obs.WithComponents(obs.JaegerComponent))).
			Setup(env.Cluster)
		obsClient = obs.From(obsDeployment, env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(ns)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh)).To(Succeed())
		Expect(env.Cluster.DeleteDeployment(obsDeployment)).To(Succeed())
	})

	It("should emit traces to jaeger", func() {
		// given TrafficTrace and mesh with tracing backend
		err := YamlK8s(meshWithTracing(mesh, obsClient.ZipkinCollectorURL()))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
		err = YamlK8s(trafficTrace(mesh, ns))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// when client sends requests to server
		clientPod, err := PodNameOfApp(env.Cluster, "demo-client", ns)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			_, _, err := env.Cluster.Exec(ns, clientPod, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "test-server")
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
