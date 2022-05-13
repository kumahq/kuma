package metrics

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	testserver "github.com/kumahq/kuma/test/framework/deployments/testserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MeshKubernetesAndMetricsAggregate(name string) InstallFunc {
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
          aggregate:
            path-stats:
              path: "/path-stats"
              port: 80
            mesh-default:
              path: "/mesh-default"
              port: 80
            service-to-override:
              path: "/service-to-override"
              port: 80
            not-working-service:
              path: "/not-working-service"
              port: 81`, name)
	return YamlK8s(mesh)
}

func MeshKubernetesAndMetricsEnabled(name string) InstallFunc {
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
          skipMTLS: true`, name)
	return YamlK8s(mesh)
}

func ApplicationsMetrics() {
	const namespace = "applications-metrics"
	const mesh = "applications-metrics"
	const meshNoAggregate = "applications-metrics-no-aggregeate"

	BeforeAll(func() {
		E2EDeferCleanup(func() {
			Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
			Expect(env.Cluster.DeleteMesh(mesh))
			Expect(env.Cluster.DeleteMesh(meshNoAggregate))
		})

		err := NewClusterSetup().
			Install(MeshKubernetesAndMetricsAggregate(mesh)).
			Install(MeshKubernetesAndMetricsEnabled(meshNoAggregate)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("test-server"),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshNoAggregate),
				testserver.WithName("test-server-dp-metrics"),
				testserver.WithPodAnnotations(map[string]string{
					"prometheus.metrics.kuma.io/aggregate-app-path":       "/my-app",
					"prometheus.metrics.kuma.io/aggregate-app-port":       "80",
					"prometheus.metrics.kuma.io/aggregate-other-app-path": "/other-app",
					"prometheus.metrics.kuma.io/aggregate-other-app-port": "80",
				}))).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("test-server-override"),
				testserver.WithPodAnnotations(map[string]string{
					"prometheus.metrics.kuma.io/aggregate-path-stats-enabled":       "false",
					"prometheus.metrics.kuma.io/aggregate-app-path":                 "/my-app",
					"prometheus.metrics.kuma.io/aggregate-app-port":                 "80",
					"prometheus.metrics.kuma.io/aggregate-service-to-override-path": "/overridden",
					"prometheus.metrics.kuma.io/aggregate-service-to-override-port": "80",
				}))).
			Setup(env.Cluster)
		Expect(err).To(Succeed())
	})

	It("should scrape metrics defined in mesh and not fail when defined service doesn't exist", func() {
		// given
		pods, err := k8s.ListPodsE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(namespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "test-server"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		clientPodName := pods[0].Name
		podIp := pods[0].Status.PodIP

		// when
		stdout, _, err := env.Cluster.Exec(namespace, clientPodName, "test-server",
			"curl", "-v", "-m", "3", "--fail", "http://"+podIp+":1234/metrics?filter=concurrency")

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

	It("should override mesh configuration by annotation", func() {
		// given
		pods, err := k8s.ListPodsE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(namespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "test-server-override"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		clientPodName := pods[0].Name
		podIp := pods[0].Status.PodIP

		// when
		stdout, _, err := env.Cluster.Exec(namespace, clientPodName, "test-server-override",
			"curl", "-v", "-m", "3", "--fail", "http://"+podIp+":1234/metrics?filter=concurrency")

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).ToNot(BeNil())

		// response doesn't exist because was disabled
		Expect(stdout).ToNot(ContainSubstring("path-stats"))
		// path has been overridden
		Expect(stdout).ToNot(ContainSubstring("service-to-override"))
		// response doesn't exist
		Expect(stdout).ToNot(ContainSubstring("not-working-service"))

		// overridden by pod
		Expect(stdout).To(ContainSubstring("overridden"))
		// overridden by pod
		Expect(stdout).To(ContainSubstring("mesh-default"))
		// added in pod
		Expect(stdout).To(ContainSubstring("my-app"))
		// metric from envoy
		Expect(stdout).To(ContainSubstring("envoy_server_concurrency"))
	})

	It("should use only configuration from dataplane", func() {
		// given
		pods, err := k8s.ListPodsE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(namespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "test-server-dp-metrics"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		clientPodName := pods[0].Name
		podIp := pods[0].Status.PodIP

		// when
		stdout, _, err := env.Cluster.Exec(namespace, clientPodName, "test-server-dp-metrics",
			"curl", "-v", "-m", "3", "--fail", "http://"+podIp+":1234/metrics?filter=concurrency")

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).ToNot(BeNil())

		// response doesn't exist because was disabled
		Expect(stdout).ToNot(ContainSubstring("path-stats"))
		// path has been overridden
		Expect(stdout).ToNot(ContainSubstring("service-to-override"))

		// overridden by pod
		Expect(stdout).To(ContainSubstring("my-app"))
		// overridden by pod
		Expect(stdout).To(ContainSubstring("other-app"))
		// metric from envoy
		Expect(stdout).To(ContainSubstring("envoy_server_concurrency"))
	})
}
