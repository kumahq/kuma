package observability

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func MeshWithMetricsTLS(mesh string) InstallFunc {
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
      tls:
        mode: providedTLS`, mesh)
	return YamlUniversal(yaml)
}

func PrometheusMetrics() {
	const mesh = "prometheus-metrics"
	const clientName = "demo-client-outside-mesh"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshWithMetricsTLS(mesh)).
			Install(TestServerUniversal("test-server", mesh,
				WithTransparentProxy(true),
				WithArgs([]string{"echo", "--instance", "test-server"}),
				WithServiceName("test-server"),
				WithDpEnvs(map[string]string{
					"KUMA_DATAPLANE_RUNTIME_METRICS_CERT_PATH": "/kuma/server.crt",
					"KUMA_DATAPLANE_RUNTIME_METRICS_KEY_PATH":  "/kuma/server.key",
				}),
			)).
			Install(TestServerUniversal("test-server-no-tls", mesh,
				WithTransparentProxy(true),
				WithArgs([]string{"echo", "--instance", "test-server-no-tls"}),
				WithServiceName("test-server-no-tls"),
			)).
			Install(MeshTrafficPermissionAllowAllUniversal(mesh)).
			Install(DemoClientUniversal(clientName, "", WithoutDataplane())).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteApp(clientName)).To(Succeed())
	})

	It("should expose metrics through https", func() {
		// given
		host := universal.Cluster.GetApp("test-server").GetIP()

		// when doing http request
		_, _, err := client.CollectResponse(
			universal.Cluster,
			clientName,
			"http://test-server.mesh:1234/metrics",
			client.Resolve("test-server.mesh:1234", fmt.Sprintf("[%s]", host)),
		)

		// then
		Expect(err).To(HaveOccurred())

		Eventually(func(g Gomega) {
			// when doing https with certs
			stdout, _, err := client.CollectResponse(
				universal.Cluster,
				clientName,
				"https://test-server.mesh:1234/metrics",
				client.WithCACert("/kuma/server.crt"),
				client.Resolve("test-server.mesh:1234", fmt.Sprintf("[%s]", host)),
			)
			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("envoy_server_live"))
		}).Should(Succeed())
	})

	It("should expose metrics through http when the certificate wasn't provided", func() {
		// given
		host := universal.Cluster.GetApp("test-server-no-tls").GetIP()

		Eventually(func(g Gomega) {
			// when
			stdout, _, err := client.CollectResponse(
				universal.Cluster,
				clientName,
				fmt.Sprintf("http://%s/metrics", net.JoinHostPort(host, "1234")),
			)

			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("envoy_server_live"))
		}).Should(Succeed())
	})
}
