package observability

import (
	"fmt"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshAndMetricsDelegated(name string) InstallFunc {
	mesh := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
spec:
  metrics:
    enabledBackend: prom-1
    backends:
      - name: prom-1
        type: prometheus
        conf:
          port: 1234
          path: /metrics
          skipMTLS: true
          tls: 
            enabled: true
            mode: delegated`, name)
	return YamlK8s(mesh)
}

func ContainerPatch(namespace string) InstallFunc {
	return YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: ContainerPatch
metadata:
  name: %s-container-patch-1
  namespace: kuma-system
spec:
  sidecarPatch:
    - op: add
      path: /env/-
      value: '{
          "name": "KUMA_DATAPLANE_METRICS_CERT_PATH",
          "value": "/kuma/server.crt"
        }'
    - op: add
      path: /env/-
      value: '{
          "name": "KUMA_DATAPLANE_METRICS_KEY_PATH",
          "value": "/kuma/server.key"
        }'`, namespace))
}

func PrometheusMetrics() {
	const namespace = "prometheus-metrics"
	const namespaceNoSidecar = "prometheus-metrics-no-sidecar"
	const mesh = "prometheus-metrics"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshAndMetricsDelegated(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(namespaceNoSidecar)).
			Install(ContainerPatch(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("test-server"),
				testserver.WithPodAnnotations(map[string]string{
					"kuma.io/container-patches": fmt.Sprintf("%s-container-patch-1", namespace),
				}),
			)).
			Install(democlient.Install(democlient.WithNamespace(namespaceNoSidecar))).
			Setup(kubernetes.Cluster)
		Expect(err).To(Succeed())
	})
	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespaceNoSidecar)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	It("should scrape metrics defined in mesh and not fail when defined service doesn't exist", func() {
		// given
		time.Sleep(3600 * time.Hour)
		podIp, err := PodIPOfApp(kubernetes.Cluster, "test-server", namespace)
		Expect(err).ToNot(HaveOccurred())

		// when
		stdout, _, err := client.CollectResponse(
			kubernetes.Cluster, "test-server", "http://"+net.JoinHostPort(podIp, "1234")+"/metrics",
			client.FromKubernetesPod(namespace, "test-server"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).ToNot(BeNil())
		// response returned by test-server
		Expect(stdout).To(ContainSubstring("path-stats"))
		// metric from envoy
		Expect(stdout).To(ContainSubstring("envoy_server_concurrency"))
		// response doesn't exist
		Expect(stdout).ToNot(ContainSubstring("not-working-service"))
	})
}
