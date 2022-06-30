package observability

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func MeshAndMetricsAggregate(name string) InstallFunc {
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
          envoy:
            filterRegex: concurrency
          skipMTLS: true
          aggregate:
          - name: path-stats
            path: "/path-stats"
            port: 80
          - name: mesh-default
            path: "/mesh-default"
            port: 80
          - name: service-to-override
            path: "/service-to-override"
            port: 80
          - name: not-working-service
            path: "/not-working-service"
            port: 81`, name)
	return YamlK8s(mesh)
}

func MeshAndMetricsEnabled(name string) InstallFunc {
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
          envoy:
            filterRegex: concurrency
          port: 1234
          path: /metrics
          skipMTLS: true`, name)
	return YamlK8s(mesh)
}

func MeshAndEnvoyMetricsFilters(name string, usedOnly string) InstallFunc {
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
          port: 5555
          path: /metrics/stats
          envoy:
            filterRegex: http2_act.*
            usedOnly: %s
          skipMTLS: true
          aggregate:
          - name: path-stats
            path: "/path-stats"
            port: 80`, name, usedOnly)
	return YamlK8s(mesh)
}

func ApplicationsMetrics() {
	const namespace = "applications-metrics"
	const mesh = "applications-metrics"
	const meshNoAggregate = "applications-metrics-no-aggregeate"
	const meshEnvoyFilter = "applications-metrics-envoy-filter"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshAndMetricsAggregate(mesh)).
			Install(MeshAndMetricsEnabled(meshNoAggregate)).
			Install(MeshAndEnvoyMetricsFilters(meshEnvoyFilter, "false")).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("test-server"),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshEnvoyFilter),
				testserver.WithName("test-server-filter"),
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
				testserver.WithName("test-server-override-mesh"),
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
	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh))
		Expect(env.Cluster.DeleteMesh(meshNoAggregate))
	})

	It("should scrape metrics defined in mesh and not fail when defined service doesn't exist", func() {
		// given
		podName, err := PodNameOfApp(env.Cluster, "test-server", namespace)
		Expect(err).ToNot(HaveOccurred())
		podIp, err := PodIPOfApp(env.Cluster, "test-server", namespace)
		Expect(err).ToNot(HaveOccurred())

		// when
		stdout, _, err := env.Cluster.Exec(namespace, podName, "test-server",
			"curl", "-v", "-m", "3", "--fail", "http://"+net.JoinHostPort(podIp, "1234")+"/metrics")

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

	It("should override mesh configuration with annotation", func() {
		// given
		podName, err := PodNameOfApp(env.Cluster, "test-server-override-mesh", namespace)
		Expect(err).ToNot(HaveOccurred())
		podIp, err := PodIPOfApp(env.Cluster, "test-server-override-mesh", namespace)
		Expect(err).ToNot(HaveOccurred())

		// when
		stdout, _, err := env.Cluster.Exec(namespace, podName, "test-server-override-mesh",
			"curl", "-v", "-m", "3", "--fail", "http://"+net.JoinHostPort(podIp, "1234")+"/metrics")

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
		podName, err := PodNameOfApp(env.Cluster, "test-server-dp-metrics", namespace)
		Expect(err).ToNot(HaveOccurred())
		podIp, err := PodIPOfApp(env.Cluster, "test-server-dp-metrics", namespace)
		Expect(err).ToNot(HaveOccurred())

		// when
		stdout, _, err := env.Cluster.Exec(namespace, podName, "test-server-dp-metrics",
			"curl", "-v", "-m", "3", "--fail", "http://"+net.JoinHostPort(podIp, "1234")+"/metrics")

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

	It("should return filtered Envoy metrics and react for change of usedOnly parameter", func() {
		// given
		podName, err := PodNameOfApp(env.Cluster, "test-server-filter", namespace)
		Expect(err).ToNot(HaveOccurred())
		podIp, err := PodIPOfApp(env.Cluster, "test-server-filter", namespace)
		Expect(err).ToNot(HaveOccurred())

		// when
		stdout, _, err := env.Cluster.Exec(namespace, podName, "test-server-filter",
			"curl", "-v", "-m", "3", "--fail", "http://"+net.JoinHostPort(podIp, "5555")+"/metrics/stats")

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).ToNot(BeNil())

		// other application metrics
		Expect(stdout).To(ContainSubstring("path-stats"))
		// metric from envoy
		Expect(stdout).To(ContainSubstring("kuma_envoy_admin"))

		// when usedOnly is enabled
		Expect(MeshAndEnvoyMetricsFilters(meshEnvoyFilter, "true")(env.Cluster)).To(Succeed())

		// then
		Eventually(func() string {
			stdout, _, err = env.Cluster.Exec(namespace, podName, "test-server-filter",
				"curl", "-v", "-m", "3", "--fail", "http://"+net.JoinHostPort(podIp, "5555")+"/metrics/stats")
			if err != nil {
				return ""
			}
			return stdout
		}, "30s", "1s").Should(And(
			ContainSubstring("path-stats"),
			Not(ContainSubstring("kuma_envoy_admin")),
		))
	})
}
