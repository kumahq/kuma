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
      - type: Zipkin
        zipkin:
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
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(ns), democlient.WithMesh(mesh)),
				testserver.Install(testserver.WithMesh(mesh), testserver.WithNamespace(ns)),
				obs.Install(obsDeployment, obs.WithNamespace(obsNs), obs.WithComponents(obs.JaegerComponent)),
			)).
			Setup(kubernetes.Cluster)
		obsClient = obs.From(obsDeployment, kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
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
		// given MeshTrace and with tracing backend
		err := YamlK8s(traceAllK8s(mesh, obsClient.ZipkinCollectorURL()))(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", "test-server",
				client.FromKubernetesPod(ns, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
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
