package externalservices

import (
	"fmt"
	"sync"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func ExternalServicesOnMultizoneUniversal() {
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
    passthrough: true
routing:
  localityAwareLoadBalancing: %s
`

	const clusterName1 = "kuma-es-1"
	const clusterName2 = "kuma-es-2"
	const clusterName3 = "kuma-es-3"
	const clusterName4 = "kuma-es-4"

	const defaultMesh = "default"

	var global, zone1, zone2, external Cluster

	BeforeEach(func() {
		// External Service non-Kuma Cluster
		external = NewUniversalCluster(NewTestingT(), clusterName4, Silent)

		// todo(lobkovilya): use test-server as an external service
		err := NewClusterSetup().
			Install(externalservice.Install(externalservice.HttpServer, externalservice.UniversalAppEchoServer)).
			Install(externalservice.Install(externalservice.HttpsServer, externalservice.UniversalAppHttpsEchoServer)).
			Install(externalservice.Install("es-for-kuma-es-2", externalservice.ExternalServiceCommand(80, "{\\\"instance\\\":\\\"kuma-es-2\\\"}"))).
			Install(externalservice.Install("es-for-kuma-es-3", externalservice.ExternalServiceCommand(80, "{\\\"instance\\\":\\\"kuma-es-3\\\"}"))).
			Setup(external)
		Expect(err).ToNot(HaveOccurred())

		externalServiceAddress := externalservice.From(external, externalservice.HttpServer).GetExternalAppAddress()
		Expect(externalServiceAddress).ToNot(BeEmpty())

		// Global
		global = NewUniversalCluster(NewTestingT(), clusterName1, Silent)
		err = NewClusterSetup().
			Install(Kuma(core.Global)).
			Install(YamlUniversal(fmt.Sprintf(meshDefaulMtlsOn, "false"))).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		wg := sync.WaitGroup{}
		wg.Add(2)
		// Cluster 1
		zone1 = NewUniversalCluster(NewTestingT(), clusterName2, Silent)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			err = NewClusterSetup().
				Install(Kuma(core.Zone,
					WithGlobalAddress(globalCP.GetKDSServerAddress()),
					WithHDS(false),
				)).
				Install(DemoClientUniversal(AppModeDemoClient, defaultMesh, WithTransparentProxy(true))).
				Setup(zone1)
			Expect(err).ToNot(HaveOccurred())
		}()

		// Cluster 2
		zone2 = NewUniversalCluster(NewTestingT(), clusterName3, Silent)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			err = NewClusterSetup().
				Install(Kuma(core.Zone,
					WithGlobalAddress(globalCP.GetKDSServerAddress()),
					WithHDS(false),
				)).
				Install(DemoClientUniversal(AppModeDemoClient, defaultMesh, WithTransparentProxy(true))).
				Setup(zone2)
			Expect(err).ToNot(HaveOccurred())
		}()
		wg.Wait()
	})

	E2EAfterEach(func() {
		Expect(external.DismissCluster()).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())
		Expect(zone2.DismissCluster()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	It("should respect external-service's zone tag in locality-aware lb mode", func() {
		externalServiceWithZone := func(zone, address string) string {
			return fmt.Sprintf(`
type: ExternalService
mesh: default
name: es-for-%s
tags:
  kuma.io/service: es-for-zones
  kuma.io/protocol: http
  kuma.io/zone: %s
networking:
  address: %s
`, zone, zone, address)
		}

		// given 2 external services with different zone tag
		Expect(YamlUniversal(externalServiceWithZone("kuma-es-2", "kuma-es-4_externalservice-es-for-kuma-es-2:80"))(global)).To(Succeed())
		Expect(YamlUniversal(externalServiceWithZone("kuma-es-3", "kuma-es-4_externalservice-es-for-kuma-es-3:80"))(global)).To(Succeed())
		// then
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(zone1, "demo-client", "es-for-zones.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(2),
				HaveKey(Equal("kuma-es-2")),
				HaveKey(Equal("kuma-es-3")),
			),
		)

		// when locality-aware lb is enabled
		Expect(YamlUniversal(fmt.Sprintf(meshDefaulMtlsOn, "true"))(global)).To(Succeed())
		// then
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(zone1, "demo-client", "es-for-zones.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(1),
				HaveKey(Equal("kuma-es-2")),
			),
		)
	})
}
