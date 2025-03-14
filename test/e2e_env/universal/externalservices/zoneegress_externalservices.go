package externalservices

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
)

func ThroughZoneEgress() {
	meshDefaulMtlsOn := `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
networking:
  outbound:
    passthrough: false
routing:
  zoneEgress: true
`

	externalService := `
type: ExternalService
mesh: default
name: external-service-1
tags:
  kuma.io/service: external-service-1
  kuma.io/protocol: http
networking:
  address: "kuma-es-ze_externalservice-http-server:80"
`

	var cluster Cluster

	const clusterName = "kuma-es-ze"

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), clusterName, Silent)

		err := NewClusterSetup().
			Install(Kuma(core.Zone)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(YamlUniversal(meshDefaulMtlsOn)).
			Install(MeshTrafficPermissionAllowAllUniversal("default")).
			Install(TestServerExternalServiceUniversal("http-server", 80, false, WithDockerContainerName("kuma-es-ze_externalservice-http-server"))).
			Install(DemoClientUniversal(AppModeDemoClient, "default", WithTransparentProxy(true))).
			Install(EgressUniversal(func(zone string) (string, error) {
				return cluster.GetKuma().GenerateZoneEgressToken("")
			})).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(cluster, "default")
	})

	E2EAfterEach(func() {
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should not access external service when zone egress is down", func() {
		// given universal cluster
		universalClusters, ok := cluster.(*UniversalCluster)
		Expect(ok).To(BeTrue())

		filter := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			"default",
			"external-service-1",
		)

		// when external service configuration is provided
		err := YamlUniversal(externalService)(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then should reach external service
		Eventually(func(g Gomega) {
			_, stderr, err := client.CollectResponse(
				cluster, "demo-client", "external-service-1.mesh",
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		}, "30s", "1s").Should(Succeed())

		// and increase stats at egress
		Eventually(func(g Gomega) {
			stat, err := cluster.GetZoneEgressEnvoyTunnel().GetStats(filter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())

		// when egress is down
		_, _, err = universalClusters.Exec("", "", AppEgress, "pkill", "kuma-dp")
		Expect(err).ToNot(HaveOccurred())
		// then traffic shouldn't reach external service
		Eventually(func(g Gomega) {
			out, _, err := universalClusters.Exec("", "", AppEgress, "pgrep", "-l", "envoy")
			Logf("output %s:", out)
			g.Expect(err).To(HaveOccurred())
			g.Expect(out).ToNot(ContainSubstring("envoy"))
		}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())

		// then traffic shouldn't reach external service
		Consistently(func(g Gomega) {
			response, err := client.CollectFailure(
				cluster, "demo-client", "external-service-1.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(22), Equal(52), Equal(56), Equal(7), Equal(28)))
		}).Should(Succeed())
	})
}
