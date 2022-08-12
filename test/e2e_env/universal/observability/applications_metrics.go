package observability

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/expfmt"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
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
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh)).To(Succeed())
		Expect(env.Cluster.DeleteMeshApps(meshNoAggregate)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshNoAggregate)).To(Succeed())
	})

	It("should scrape metrics defined in mesh and not fail when defined service doesn't exist", func() {
		// given
		ip := env.Cluster.GetApp("test-server").GetIP()

		// when
		stdout, _, err := env.Cluster.ExecWithRetries("", "", "test-server",
			"curl", "-v", "-m", "3", "--fail", "http://"+net.JoinHostPort(ip, "1234")+"/metrics?filter=concurrency")

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

	It("should override mesh configuration with dataplane configuration", func() {
		// given
		ip := env.Cluster.GetApp("test-server-override-mesh").GetIP()

		// when
		stdout, _, err := env.Cluster.ExecWithRetries("", "", "test-server-override-mesh",
			"curl", "-v", "-m", "3", "--fail", "http://"+net.JoinHostPort(ip, "1234")+"/metrics/overridden?filter=concurrency")

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
		ip := env.Cluster.GetApp("test-server-dp-metrics").GetIP()

		// when
		stdout, _, err := env.Cluster.ExecWithRetries("", "", "test-server-dp-metrics",
			"curl", "-v", "-m", "3", "--fail", "http://"+net.JoinHostPort(ip, "5555")+"/stats?filter=concurrency")

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).ToNot(BeNil())
		Expect(stdout).To(ContainSubstring(string(expfmt.FmtText)))

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
