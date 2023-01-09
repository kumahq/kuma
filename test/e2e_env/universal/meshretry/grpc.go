package meshretry

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
)

func GrpcRetry() {
	meshName := "meshretry-grpc"
	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(TestServerUniversal("test-server", meshName,
				WithServiceName("test-server"),
				WithProtocol("grpc"),
				WithArgs([]string{"grpc", "server", "--port", "8080"}),
				WithTransparentProxy(true),
			)).
			Install(TestServerUniversal("test-client", meshName,
				WithServiceName("test-client"),
				WithArgs([]string{"grpc", "client", "--address", "test-server.mesh:80", "--unary", "true"}),
				WithProtocol("grpc"),
				WithTransparentProxy(true),
			)).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Delete the default retry policy
		Eventually(func() error {
			return env.Cluster.GetKumactlOptions().RunKumactl("delete", "retry", "--mesh", meshName, "retry-all-"+meshName)
		}).Should(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should retry on GRPC connection failure", func() {
		echoServerDataplane := fmt.Sprintf(`
type: Dataplane
mesh: "%s"
name: fake-echo-server
networking:
  address:  241.0.0.1
  inbound:
  - port: 7777
    servicePort: 7777
    tags:
      kuma.io/service: test-server
      kuma.io/protocol: grpc
`, meshName)
		meshRetryPolicy := fmt.Sprintf(`
type: MeshRetry
mesh: "%s"
name: fake-meshretry-policy
spec:
  targetRef:
    kind: MeshService
    name: test-client
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        grpc:
          numRetries: 5
`, meshName)
		admin, err := env.Cluster.GetApp("test-client").GetEnvoyAdminTunnel()
		Expect(err).ToNot(HaveOccurred())

		var lastFailureStats = stats.StatItem{Name: "", Value: float64(0)}
		grpcFailureStats := func(g Gomega) *stats.Stats {
			s, err := admin.GetStats("cluster.test-server.grpc.failure")
			g.Expect(err).ToNot(HaveOccurred())
			fmt.Printf("current failure stats %v\n", s)
			return s
		}
		grpcSuccessStats := func(g Gomega) *stats.Stats {
			s, err := admin.GetStats("cluster.test-server.grpc.success")
			g.Expect(err).ToNot(HaveOccurred())
			return s
		}

		By("Checking requests succeed")
		Eventually(func(g Gomega) {
			g.Expect(grpcSuccessStats(g)).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())

		Consistently(func(g Gomega) {
			failureStats := grpcFailureStats(g)
			if len(failureStats.Stats) != 0 {
				defer func() { lastFailureStats = failureStats.Stats[0] }()
				g.Expect(failureStats).To(Not(stats.BeGreaterThan(lastFailureStats)))
			}
		}).Should(Succeed())

		By("Adding a faulty dataplane")
		Expect(env.Cluster.Install(YamlUniversal(echoServerDataplane))).To(Succeed())

		By("Check some errors happen")
		Eventually(func(g Gomega) {
			failureStats := grpcFailureStats(g)
			defer func() { lastFailureStats = failureStats.Stats[0] }()
			g.Expect(grpcFailureStats(g)).To(stats.BeGreaterThanZero())
		}, "90s", "10s").Should(Succeed())
		Consistently(func(g Gomega) {
			failureStats := grpcFailureStats(g)
			defer func() { lastFailureStats = failureStats.Stats[0] }()
			g.Expect(failureStats).To(stats.BeGreaterThanZero())
			g.Expect(failureStats).To(stats.BeGreaterThan(lastFailureStats))
		}, "40s", "10s").Should(Succeed())

		By("Apply a MeshRetry policy")
		Expect(env.Cluster.Install(YamlUniversal(meshRetryPolicy))).To(Succeed())

		By("Eventually all requests succeed consistently")
		Eventually(func(g Gomega) {
			failureStats := grpcFailureStats(g)
			defer func() { lastFailureStats = failureStats.Stats[0] }()
			g.Expect(failureStats).To(Not(stats.BeGreaterThan(lastFailureStats)))
		}, "50s", "10s").Should(Succeed())
		Consistently(func(g Gomega) {
			failureStats := grpcFailureStats(g)
			defer func() { lastFailureStats = failureStats.Stats[0] }()
			g.Expect(failureStats).To(Not(stats.BeGreaterThan(lastFailureStats)))
		}, "60s", "10s").Should(Succeed())
	})
}
