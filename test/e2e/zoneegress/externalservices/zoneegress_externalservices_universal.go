package externalservices

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
)

func UniversalStandalone() {
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
  address: "kuma-3_externalservice-http-server:80"
`

	var cluster Cluster

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma3},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		cluster = clusters.GetCluster(Kuma3)

		err = NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(YamlUniversal(meshDefaulMtlsOn)).
			Install(externalservice.Install(externalservice.HttpServer, externalservice.UniversalAppEchoServer)).
			Install(DemoClientUniversal(AppModeDemoClient, "default", WithTransparentProxy(true))).
			Install(EgressUniversal(func(zone string) (string, error) {
				return cluster.GetKuma().GenerateZoneEgressToken("")
			})).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	serviceUnreachable := func() error {
		_, _, err := cluster.Exec("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "external-service-1.mesh")
		return err
	}

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
		stdout, _, err := cluster.ExecWithRetries("", "", "demo-client",
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-1.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		// and increase stats at egress
		Eventually(func(g Gomega) {
			stat, err := cluster.GetZoneEgressEnvoyTunnel().GetStats(filter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())

		// when egress is down
		_, _, err = universalClusters.Exec("", "", AppEgress, "pkill", "-9", "kuma-dp")
		Expect(err).ToNot(HaveOccurred())

		// then traffic shouldn't reach external service
		Eventually(serviceUnreachable, "30s", "1s").Should(HaveOccurred())
	})
}
