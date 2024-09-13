package delegated

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/observability"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshTrace(config *Config) func() {
	GinkgoHelper()

	return func() {
		var observabilityClient observability.Observability

		meshTrace := func(zipkinUrl string) framework.InstallFunc {
			return framework.YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrace
metadata:
  name: trace-all-delegated
  namespace: %s
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
`, config.CpNamespace, config.Mesh, zipkinUrl))
		}

		framework.AfterEachFailure(func() {
			framework.DebugKube(kubernetes.Cluster, config.Mesh, config.Namespace, config.ObservabilityDeploymentName)
		})

		framework.E2EAfterEach(func() {
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				v1alpha1.MeshTraceResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should emit traces to jaeger", func() {
			observabilityClient = observability.From(config.ObservabilityDeploymentName, kubernetes.Cluster)

			// given MeshTrace and with tracing backend
			Expect(kubernetes.Cluster.Install(meshTrace(observabilityClient.ZipkinCollectorURL()))).
				To(Succeed())

			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				srvs, err := observabilityClient.TracedServices()
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(srvs).To(Equal([]string{
					fmt.Sprintf("delegated-gateway-admin_%s_svc_8444", config.Mesh),
					"jaeger-query",
					fmt.Sprintf("test-server_%s_svc_80", config.Mesh),
				}))
			}, "30s", "1s").Should(Succeed())
		})
	}
}
