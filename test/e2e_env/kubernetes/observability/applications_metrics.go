package observability

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
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
			Install(Parallel(
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(mesh),
					testserver.WithName("test-server"),
				),
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshEnvoyFilter),
					testserver.WithName("test-server-filter"),
				),
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshNoAggregate),
					testserver.WithName("test-server-dp-metrics"),
					testserver.WithoutProbes(), // when application binds to localhost you cannot access it
					testserver.WithEchoArgs("--ip", "localhost"),
					testserver.WithPodAnnotations(map[string]string{
						"prometheus.metrics.kuma.io/aggregate-app-path":       "/my-app",
						"prometheus.metrics.kuma.io/aggregate-app-port":       "80",
						"prometheus.metrics.kuma.io/aggregate-app-address":    "localhost",
						"prometheus.metrics.kuma.io/aggregate-other-app-path": "/other-app",
						"prometheus.metrics.kuma.io/aggregate-other-app-port": "80",
					})),
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(mesh),
					testserver.WithName("test-server-override-mesh"),
					testserver.WithPodAnnotations(map[string]string{
						"prometheus.metrics.kuma.io/aggregate-path-stats-enabled":       "false",
						"prometheus.metrics.kuma.io/aggregate-app-path":                 "/my-app",
						"prometheus.metrics.kuma.io/aggregate-app-port":                 "80",
						"prometheus.metrics.kuma.io/aggregate-service-to-override-path": "/overridden",
						"prometheus.metrics.kuma.io/aggregate-service-to-override-port": "80",
					})),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, mesh, namespace)
		DebugKube(kubernetes.Cluster, meshNoAggregate)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshNoAggregate)).To(Succeed())
	})

	It("should scrape metrics defined in mesh and not fail when defined service doesn't exist", func() {
		// given
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

	It("should override mesh configuration with annotation", func() {
		// given
		podIp, err := PodIPOfApp(kubernetes.Cluster, "test-server-override-mesh", namespace)
		Expect(err).ToNot(HaveOccurred())

		// when
		stdout, _, err := client.CollectResponse(
			kubernetes.Cluster, "test-server-override-mesh", "http://"+net.JoinHostPort(podIp, "1234")+"/metrics",
			client.FromKubernetesPod(namespace, "test-server-override-mesh"),
		)

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
		podIp, err := PodIPOfApp(kubernetes.Cluster, "test-server-dp-metrics", namespace)
		Expect(err).ToNot(HaveOccurred())

		// when
		stdout, _, err := client.CollectResponse(
			kubernetes.Cluster, "test-server-dp-metrics", "http://"+net.JoinHostPort(podIp, "1234")+"/metrics",
			client.FromKubernetesPod(namespace, "test-server-dp-metrics"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).ToNot(BeNil())

		// response doesn't exist because was disabled
		Expect(stdout).ToNot(ContainSubstring("path-stats"))
		// path has been overridden
		Expect(stdout).ToNot(ContainSubstring("service-to-override"))

		// overridden by pod
		Expect(stdout).To(ContainSubstring("my-app"))
		// overridden by pod but binding to localhost so it's not available
		Expect(stdout).ToNot(ContainSubstring("other-app"))
		// metric from envoy
		Expect(stdout).To(ContainSubstring("envoy_server_concurrency"))
	})

	It("should return filtered Envoy metrics and react for change of usedOnly parameter", func() {
		// given
		podIp, err := PodIPOfApp(kubernetes.Cluster, "test-server-filter", namespace)
		Expect(err).ToNot(HaveOccurred())

		// when
		stdout, _, err := client.CollectResponse(
			kubernetes.Cluster, "test-server-filter", "http://"+net.JoinHostPort(podIp, "5555")+"/metrics/stats",
			client.FromKubernetesPod(namespace, "test-server-filter"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).ToNot(BeNil())

		// other application metrics
		Expect(stdout).To(ContainSubstring("path-stats"))
		// metric from envoy
		Expect(stdout).To(ContainSubstring("kuma_envoy_admin"))

		// when usedOnly is enabled
		Expect(MeshAndEnvoyMetricsFilters(meshEnvoyFilter, "true")(kubernetes.Cluster)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				kubernetes.Cluster, "test-server-filter", "http://"+net.JoinHostPort(podIp, "5555")+"/metrics/stats",
				client.FromKubernetesPod(namespace, "test-server-filter"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(And(
				ContainSubstring("path-stats"),
				Not(ContainSubstring("kuma_envoy_admin")),
			))
		}, "30s", "1s").Should(Succeed())
	})
}
