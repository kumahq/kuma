package delegated

import (
	"fmt"
	"net"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshopentelemetrybackend_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/otelcollector"
	"github.com/kumahq/kuma/v3/test/framework/envs/kubernetes"
)

func MeshMetric(config *Config) func() {
	GinkgoHelper()

	return func() {
		otelBackend := func(otelEndpoint string) framework.InstallFunc {
			host, port, err := net.SplitHostPort(otelEndpoint)
			Expect(err).ToNot(HaveOccurred())
			return framework.YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshOpenTelemetryBackend
metadata:
  name: otel-backend-delegated
  namespace: %s
  labels:
    kuma.io/mesh: %s
    kuma.io/display-name: otel-backend-delegated
spec:
  endpoint:
    address: %s
    port: %s
`, config.CpNamespace, config.Mesh, host, port))
		}

		meshMetric := func(otelEndpoint string) framework.InstallFunc {
			return framework.Combine(
				otelBackend(otelEndpoint),
				framework.YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshMetric
metadata:
  name: otel-metrics-delegated
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  default:
    sidecar:
      profiles:
        appendProfiles:
          - name: All
    backends:
      - type: OpenTelemetry
        openTelemetry:
          backendRef:
            kind: MeshOpenTelemetryBackend
            labels:
              kuma.io/display-name: otel-backend-delegated
          refreshInterval: 30s
`, config.CpNamespace, config.Mesh)),
			)
		}

		framework.AfterEachFailure(func() {
			framework.DebugKube(kubernetes.Cluster, config.Mesh, config.Namespace, config.ObservabilityDeploymentName)
		})

		framework.E2EAfterEach(func() {
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				v1alpha1.MeshMetricResourceTypeDescriptor,
			)).To(Succeed())
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				meshopentelemetrybackend_api.MeshOpenTelemetryBackendResourceTypeDescriptor,
			)).To(Succeed())
		})

		XIt("MeshMetric with OpenTelemetry enabled", func() {
			// given
			collector := otelcollector.From(kubernetes.Cluster, otelcollector.DefaultDeploymentName)
			Expect(kubernetes.Cluster.Install(meshMetric(collector.CollectorEndpoint()))).To(Succeed())

			// then
			Eventually(func(g Gomega) {
				stdout, _, err := client.CollectResponse(
					kubernetes.Cluster,
					"demo-client",
					collector.ExporterEndpoint(),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
					client.WithVerbose(),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stdout).To(WithTransform(
					func(stdout string) []string {
						return strings.Split(stdout, "\n")
					},
					ContainElement(MatchRegexp(
						`envoy_cluster_external_upstream_rq_time_bucket\{.*service="%[1]s-gateway-admin_%[1]s_svc_8444"`,
						config.Mesh,
					)),
				))
			}, "3m", "5s").Should(Succeed())
		})
	}
}
