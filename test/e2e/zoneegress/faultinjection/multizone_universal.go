package faultinjection

import (
	"fmt"
	"net"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func meshMTLSOn(mesh string) string {
	return fmt.Sprintf(`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`, mesh)
}

func ExternalServerUniversal(name string) InstallFunc {
	return func(cluster Cluster) error {
		return cluster.DeployApp(
			WithArgs([]string{"test-server", "echo", "--port", "8080", "--instance", name}),
			WithName(name),
			WithoutDataplane(),
			WithVerbose())
	}
}

var externalServiceV1 = `
type: ExternalService
mesh: %s
name: external-service-v1
tags:
  kuma.io/service: external-service
  kuma.io/protocol: http
  version: v1
networking:
  address: "%s"`

func ExternalServiceV1(mesh, address string) string {
	return fmt.Sprintf(externalServiceV1, mesh, address)
}

var externalServiceV2 = `
type: ExternalService
mesh: %s
name: external-service-v2
tags:
  kuma.io/service: external-service
  kuma.io/protocol: http
  version: v2
networking:
  address: "%s"`

func ExternalServiceV2(mesh, address string) string {
	return fmt.Sprintf(externalServiceV2, mesh, address)
}

var global, universalZone *UniversalCluster

var _ = E2EBeforeSuite(func() {

	universalClusters, err := NewUniversalClusters(
		[]string{Kuma4, Kuma5},
		Silent)
	Expect(err).ToNot(HaveOccurred())

	// Global
	global = universalClusters.GetCluster(Kuma5).(*UniversalCluster)
	Expect(NewClusterSetup().
		Install(Kuma(config_core.Global)).
		Install(YamlUniversal(meshMTLSOn("default"))).
		Setup(global)).To(Succeed())

	demoClientToken, err := global.GetKuma().GenerateDpToken("default", "dp-demo-client")
	Expect(err).ToNot(HaveOccurred())

	egressTokenZone4, err := global.GetKuma().GenerateZoneEgressToken(Kuma4)
	Expect(err).ToNot(HaveOccurred())

	testServerToken, err := global.GetKuma().GenerateDpToken("default", "test-server")
	Expect(err).ToNot(HaveOccurred())

	// Universal Zone
	universalZone = universalClusters.GetCluster(Kuma4).(*UniversalCluster)
	Expect(NewClusterSetup().
		Install(Kuma(config_core.Zone, WithGlobalAddress(global.GetKuma().GetKDSServerAddress()))).
		Install(DemoClientUniversal("dp-demo-client", "default", demoClientToken, WithTransparentProxy(true))).
		Install(TestServerUniversal("test-server", "default", testServerToken, WithArgs([]string{"echo", "--instance", "universal1"}))).
		Install(EgressUniversal(egressTokenZone4)).
		Install(ExternalServerUniversal("es-test-server-v1")).
		Install(ExternalServerUniversal("es-test-server-v2")).
		Setup(universalZone)).To(Succeed())

	Expect(global.GetKumactlOptions().
		KumactlApplyFromString(
			ExternalServiceV1(
				"default",
				net.JoinHostPort(universalZone.GetApp("es-test-server-v1").GetIP(), "8080"),
			)),
	).To(Succeed())

	Expect(global.GetKumactlOptions().
		KumactlApplyFromString(
			ExternalServiceV2(
				"default",
				net.JoinHostPort(universalZone.GetApp("es-test-server-v2").GetIP(), "8080"),
			)),
	).To(Succeed())

	E2EDeferCleanup(func() {
		Expect(global.DeleteKuma()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())

		Expect(universalZone.DeleteKuma()).To(Succeed())
		Expect(universalZone.DismissCluster()).To(Succeed())
	})
})

func MultizoneUniversal() {
	E2EAfterEach(func() {
		// remove all FaultInjections
		items, err := global.GetKumactlOptions().KumactlList("fault-injections", "default")
		Expect(err).ToNot(HaveOccurred())
		for _, item := range items {
			err := global.GetKumactlOptions().KumactlDelete("fault-injection", item, "default")
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("should inject faults for external service", func() {
		Expect(YamlUniversal(`
type: FaultInjection
mesh: default
name: fi1
sources:
   - match:
       kuma.io/service: dp-demo-client
destinations:
   - match:
       kuma.io/service: external-service
       kuma.io/protocol: http
       version: v2
conf:
   abort:
     httpStatus: 401
     percentage: 100`)(global)).To(Succeed())

		Eventually(func() bool {
			stdout, _, err := universalZone.Exec("", "", "dp-demo-client",
				"curl", "-v", "-m", "8", "external-service.mesh")
			if err != nil {
				return false
			}
			return strings.Contains(stdout, "401 Unauthorized")
		}, "30s", "1s").Should(BeTrue())
	})
}
