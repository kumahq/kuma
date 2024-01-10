package meshretry

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/envs/universal"
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
			// remove default policies after https://github.com/kumahq/kuma/issues/3325
			Install(TrafficRouteUniversal(meshName)).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Delete the default meshretry policy
		Eventually(func() error {
			return universal.Cluster.GetKumactlOptions().RunKumactl("delete", "meshretry", "--mesh", meshName, "mesh-retry-all-"+meshName)
		}).Should(Succeed())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.GetKumactlOptions().RunKumactl("delete", "dataplane", "fake-echo-server", "-m", meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
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
		admin, err := universal.Cluster.GetApp("test-client").GetEnvoyAdminTunnel()
		Expect(err).ToNot(HaveOccurred())

		lastFailureStats := stats.StatItem{Name: "", Value: float64(0)}
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
		Expect(universal.Cluster.Install(YamlUniversal(echoServerDataplane))).To(Succeed())

		By("Check some errors happen")
		Eventually(func(g Gomega) {
			failureStats := grpcFailureStats(g)
			defer func() { lastFailureStats = failureStats.Stats[0] }()
			g.Expect(grpcFailureStats(g)).To(stats.BeGreaterThanZero())
		}, "90s", "10s").Should(Succeed())
		time.Sleep(10 * time.Second)
		Consistently(func(g Gomega) {
			failureStats := grpcFailureStats(g)
			defer func() { lastFailureStats = failureStats.Stats[0] }()
			g.Expect(failureStats).To(stats.BeGreaterThanZero())
			g.Expect(failureStats).To(stats.BeGreaterThan(lastFailureStats))
		}, "40s", "10s").Should(Succeed())

		By("Apply a MeshRetry policy")
		Expect(universal.Cluster.Install(YamlUniversal(meshRetryPolicy))).To(Succeed())

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
