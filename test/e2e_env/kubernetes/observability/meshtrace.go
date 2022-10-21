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

func traceAllK8s(meshName string, url string) string {
	return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrace
metadata:
  name: trace-all
  namespace: kuma-system
  labels:
    kuma.io/mesh: %s
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
	ns := "meshtrace"
	obsNs := "obs-meshtrace"
	obsDeployment := "obs-trace-deployment"
	mesh := "meshtrace"

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
		// given MeshTrace and with tracing backend
		err := YamlK8s(traceAllK8s(mesh, obsClient.ZipkinCollectorURL()))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// when client sends requests to server
		clientPod, err := PodNameOfApp(env.Cluster, "demo-client", ns)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() ([]string, error) {
			_, _, err := env.Cluster.ExecWithRetries(ns, clientPod, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "test-server")
			if err != nil {
				return nil, err
			}
			// then traces are published
			return obsClient.TracedServices()
		}, "30s", "1s").Should(Equal([]string{
			fmt.Sprintf("demo-client_%s_svc", ns),
			"jaeger-query",
			fmt.Sprintf("test-server_%s_svc_80", ns),
		}))
	})
}
