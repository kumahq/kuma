package grpc

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/envoy_admin/tunnel"
	"github.com/kumahq/kuma/v3/test/framework/envs/universal"
)

func GRPC() {
	meshName := "grpc"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(TestServerUniversal(
				"test-server", meshName,
				WithServiceName("test-server"),
				WithProtocol("grpc"),
				WithArgs([]string{"grpc", "server", "--port", "8080"}),
				WithTransparentProxy(true),
			)).
			Install(TestServerUniversal(
				"second-test-server", meshName,
				WithServiceName("second-test-server"),
				WithProtocol("grpc"),
				WithArgs([]string{"grpc", "server", "--port", "8080"}),
				WithTransparentProxy(true),
			)).
			Install(TestServerUniversal(
				"test-client", meshName,
				WithServiceName("test-client"),
				WithArgs([]string{"grpc", "client", "--unary", "--address", "test-server.svc.mesh.local:80"}),
				WithTransparentProxy(true),
				WithLabels(map[string]string{"kuma.io/service": "test-client"}),
			)).
			Setup(universal.Cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should emit stats from the server", func() {
		Eventually(func(g Gomega) {
			cmd := tunnel.AdminCurlCmd("/stats?format=prometheus")
			stdout, _, err := universal.Cluster.Exec("", "", "test-server", cmd...)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring(`envoy_cluster_grpc_request_message_count{envoy_cluster_name="self_inbound_dp_80"}`))
			g.Expect(stdout).To(ContainSubstring(`envoy_cluster_grpc_response_message_count{envoy_cluster_name="self_inbound_dp_80"}`))
		}, "30s", "1s").Should(Succeed())
	})

	It("should emit stats from the client", func() {
		Eventually(func(g Gomega) {
			cmd := tunnel.AdminCurlCmd("/stats?format=prometheus")
			stdout, _, err := universal.Cluster.Exec("", "", "test-client", cmd...)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring(`envoy_cluster_grpc_request_message_count{envoy_cluster_name="kri_msvc_grpc_kuma-3__test-server_80"}`))
			g.Expect(stdout).To(ContainSubstring(`envoy_cluster_grpc_response_message_count{envoy_cluster_name="kri_msvc_grpc_kuma-3__test-server_80"}`))
		}, "30s", "1s").Should(Succeed())
	})

	// todo switch to MeshGRPCRoute when https://github.com/kumahq/kuma/issues/3325 is implemented
	It("MeshHTTPRoute does split the traffic between two gRPC services", func() {
		yaml := `
type: MeshHTTPRoute
name: http-route-1
mesh: grpc
spec:
  targetRef:
    kind: Dataplane
    labels:
      kuma.io/service: test-client
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /
          default:
            backendRefs:
              - kind: MeshService
                name: test-server
                weight: 50
              - kind: MeshService
                name: second-test-server
                weight: 50
`
		Expect(universal.Cluster.Install(YamlUniversal(yaml))).To(Succeed())

		Eventually(func(g Gomega) {
			cmd := tunnel.AdminCurlCmd("/stats?format=prometheus")
			stdout, _, err := universal.Cluster.Exec("", "", "second-test-server", cmd...)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(stdout).To(ContainSubstring(`envoy_cluster_grpc_request_message_count{envoy_cluster_name="self_inbound_dp_80"}`))
			g.Expect(stdout).To(ContainSubstring(`envoy_cluster_grpc_response_message_count{envoy_cluster_name="self_inbound_dp_80"}`))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())
	})
}
