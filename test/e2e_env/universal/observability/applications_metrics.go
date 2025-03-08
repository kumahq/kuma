package observability

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/expfmt"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func MeshWithMetricsEnabeld(mesh string) InstallFunc {
	yaml := fmt.Sprintf(`
type: Mesh
name: %s
metrics:
  enabledBackend: prometheus-1
  backends:
  - name: prometheus-1
    type: prometheus
    conf:
      path: /metrics
      port: 1234`, mesh)
	return YamlUniversal(yaml)
}

func MeshKubernetesAndMetricsAggregate(mesh string) InstallFunc {
	yaml := fmt.Sprintf(`
type: Mesh
name: %s
metrics:
  enabledBackend: prometheus-1
  backends:
  - name: prometheus-1
    type: prometheus
    conf:
      path: /metrics
      port: 1234
      aggregate:
      - name: path-stats
        path: "/path-stats"
        port: 8080
      - name: mesh-default
        path: "/mesh-default"
        port: 8080
      - name: service-to-override
        path: "/service-to-override"
        port: 8080
      - name: not-working-service
        path: "/not-working-service"
        port: 81`, mesh)
	return YamlUniversal(yaml)
}

func ApplicationsMetrics() {
	const mesh = "applications-metrics"
	const meshNoAggregate = "applications-metrics-no-aggregate"
	const meshWithLocalhostBound = "applications-metrics-localhost-bound"

	dpLocalhostBoundAggregateConfig := `
metrics:
  type: prometheus
  conf:
    aggregate:
    - name: localhost-bound-not-exposed
      port: 8080
      path: "/localhost-bound-not-exposed"
    - name: localhost-bound
      port: 8080
      address: "127.0.0.1"
      path: "/localhost-bound"`

	dpAggregateConfig := `
metrics:
  type: prometheus
  conf:
    path: /stats
    port: 5555
    aggregate:
    - name: app
      path: "/my-app"
      port: 8080
    - name: other-app
      path: "/other-app"
      port: 8080`

	dpOverrideMeshAggregateConfig := `
metrics:
  type: prometheus
  conf:
    path: /metrics/overridden
    aggregate:
    - name: path-stats
      enabled: false
    - name: app
      path: "/my-app"
      port: 8080
    - name: service-to-override
      path: "/overridden"
      port: 8080`

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshWithMetricsEnabeld(meshNoAggregate)).
			Install(MeshKubernetesAndMetricsAggregate(mesh)).
			Install(MeshWithMetricsEnabeld(meshWithLocalhostBound)).
			Install(TestServerUniversal("test-server", mesh,
				WithTransparentProxy(true),
				WithArgs([]string{"echo", "--instance", "test-server"}),
				WithServiceName("test-server"),
			)).
			Install(TestServerUniversal("test-server-override-mesh", mesh,
				WithTransparentProxy(true),
				WithArgs([]string{"echo", "--instance", "test-server-override-mesh"}),
				WithServiceName("test-server-override-mesh"),
				WithAppendDataplaneYaml(dpOverrideMeshAggregateConfig),
			)).
			Install(TestServerUniversal("test-server-dp-metrics", meshNoAggregate,
				WithTransparentProxy(true),
				WithArgs([]string{"echo", "--instance", "test-server-dp-metrics"}),
				BoundToContainerIp(),
				WithServiceName("test-server-dp-metrics"),
				WithAppendDataplaneYaml(dpAggregateConfig),
			)).
			Install(TestServerUniversal("test-server-dp-metrics-localhost", meshWithLocalhostBound,
				WithTransparentProxy(true),
				WithArgs([]string{"echo", "--instance", "test-server-dp-metrics-localhost", "--ip", "127.0.0.1"}),
				WithServiceName("test-server-dp-metrics-localhost"),
				WithAppendDataplaneYaml(dpLocalhostBoundAggregateConfig),
			)).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, mesh)
		DebugUniversal(universal.Cluster, meshNoAggregate)
		DebugUniversal(universal.Cluster, meshWithLocalhostBound)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMeshApps(meshNoAggregate)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshNoAggregate)).To(Succeed())
		Expect(universal.Cluster.DeleteMeshApps(meshWithLocalhostBound)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshWithLocalhostBound)).To(Succeed())
	})

	It("should scrape metrics defined in mesh and not fail when defined service doesn't exist", func() {
		// given
		ip := universal.Cluster.GetApp("test-server").GetIP()

		// when
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				universal.Cluster, "test-server", "http://"+net.JoinHostPort(ip, "1234")+"/metrics?filter=concurrency",
			)

			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).ToNot(BeNil())
			// response returned by test-server
			g.Expect(stdout).To(ContainSubstring("path-stats"))
			// metric from envoy
			g.Expect(stdout).To(ContainSubstring("envoy_server_concurrency"))
			// response doesn't exist
			g.Expect(stdout).ToNot(ContainSubstring("not-working-service"))
		}).Should(Succeed())
	})

	It("should override mesh configuration with dataplane configuration", func() {
		// given
		ip := universal.Cluster.GetApp("test-server-override-mesh").GetIP()

		// when
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				universal.Cluster, "test-server-override-mesh", "http://"+net.JoinHostPort(ip, "1234")+"/metrics/overridden?filter=concurrency",
			)

			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).ToNot(BeNil())

			// response doesn't exist because was disabled
			g.Expect(stdout).ToNot(ContainSubstring("path-stats"))
			// path has been overridden
			g.Expect(stdout).ToNot(ContainSubstring("service-to-override"))
			// response doesn't exist
			g.Expect(stdout).ToNot(ContainSubstring("not-working-service"))

			// overridden by pod
			g.Expect(stdout).To(ContainSubstring("overridden"))
			// overridden by pod
			g.Expect(stdout).To(ContainSubstring("mesh-default"))
			// added in pod
			g.Expect(stdout).To(ContainSubstring("my-app"))
			// metric from envoy
			g.Expect(stdout).To(ContainSubstring("envoy_server_concurrency"))
		}).Should(Succeed())
	})

	It("should use only configuration from dataplane", func() {
		// given
		ip := universal.Cluster.GetApp("test-server-dp-metrics").GetIP()

		// when
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				universal.Cluster, "test-server-dp-metrics", "http://"+net.JoinHostPort(ip, "5555")+"/stats?filter=concurrency",
				client.WithHeader("Accept", "text/plain; version=0.0.4; charset=utf-8"),
				client.WithVerbose(),
			)

			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).ToNot(BeNil())
			g.Expect(stdout).To(ContainSubstring(string(expfmt.NewFormat(expfmt.TypeTextPlain))))

			// response doesn't exist because was disabled
			g.Expect(stdout).ToNot(ContainSubstring("path-stats"))
			// path has been overridden
			g.Expect(stdout).ToNot(ContainSubstring("service-to-override"))

			// overridden by pod
			g.Expect(stdout).To(ContainSubstring("my-app"))
			// overridden by pod
			g.Expect(stdout).To(ContainSubstring("other-app"))
			// metric from envoy
			g.Expect(stdout).To(ContainSubstring("envoy_server_concurrency"))
		}).Should(Succeed())
	})

	It("should allow to define and expose localhost bound server", func() {
		// given
		ip := universal.Cluster.GetApp("test-server-dp-metrics-localhost").GetIP()

		// when
		Eventually(func(g Gomega) {
			stdout, stderr, err := client.CollectResponse(
				universal.Cluster, "test-server-dp-metrics-localhost", "http://"+net.JoinHostPort(ip, "1234")+"/metrics?filter=concurrency",
				client.WithHeader("Accept", "application/openmetrics-text;version=1.0.0,application/openmetrics-text;version=0.0.1;q=0.75,text/plain;version=0.0.4;q=0.5,*/*;q=0.1"),
				client.WithVerbose(),
			)

			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).ToNot(BeNil())
			g.Expect(stderr).To(ContainSubstring(string(expfmt.NewFormat(expfmt.TypeTextPlain))))

			// path doesn't have defined address
			g.Expect(stdout).ToNot(ContainSubstring("localhost-bound-not-exposed"))

			// exposed localhost bound service
			g.Expect(stdout).To(ContainSubstring("localhost-bound"))
			// metric from envoy
			g.Expect(stdout).To(ContainSubstring("envoy_server_concurrency"))
		}).Should(Succeed())
	})
}
