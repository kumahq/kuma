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
`

	externalService := `
type: ExternalService
mesh: default
name: external-service-%s
tags:
  kuma.io/service: external-service-%s
  kuma.io/protocol: http
networking:
  address: %s
  tls:
    enabled: %s
    caCert:
      inline: "%s"
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

		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		egressToken, err := cluster.GetKuma().GenerateZoneEgressToken("")
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(externalservice.Install(externalservice.HttpServer, externalservice.UniversalAppEchoServer)).
			Install(DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithTransparentProxy(true))).
			Install(EgressUniversal(egressToken)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(meshDefaulMtlsOn)(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		err := cluster.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should access external service through zoneegress", func() {
		filter := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			"default",
			"external-service-1",
		)

		err := YamlUniversal(fmt.Sprintf(externalService,
			"1", "1",
			"kuma-3_externalservice-http-server:80",
			"false", ""))(cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			stat, err := cluster.GetZoneEgressEnvoyTunnel().GetStats(filter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "30s", "1s").Should(Succeed())

		stdout, _, err := cluster.ExecWithRetries("", "", "demo-client",
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-1.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(ContainSubstring("Echo 80"))

		Eventually(func(g Gomega) {
			stat, err := cluster.GetZoneEgressEnvoyTunnel().GetStats(filter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())
	})
}
