package faultinjection

import (
	"fmt"
	"net"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
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
routing:
  zoneEgress: true
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

var cluster *UniversalCluster

var _ = E2EBeforeSuite(func() {

	clusters, err := NewUniversalClusters(
		[]string{Kuma3},
		Silent)
	Expect(err).ToNot(HaveOccurred())

	cluster = clusters.GetCluster(Kuma3).(*UniversalCluster)

	err = NewClusterSetup().
		Install(Kuma(config_core.Standalone)).
		Setup(cluster)
	Expect(err).ToNot(HaveOccurred())

	err = NewClusterSetup().
		Install(externalservice.Install(externalservice.HttpServer, externalservice.UniversalAppEchoServer)).
		Install(DemoClientUniversal(AppModeDemoClient, "default", WithTransparentProxy(true))).
		Install(EgressUniversal(cluster.GetKuma().GenerateZoneEgressToken)).
		Install(YamlUniversal(meshMTLSOn("default"))).
		Install(ExternalServerUniversal("es-test-server-v1")).
		Install(ExternalServerUniversal("es-test-server-v2")).
		Setup(cluster)
	Expect(err).ToNot(HaveOccurred())

	Expect(cluster.GetKumactlOptions().
		KumactlApplyFromString(
			ExternalServiceV1(
				"default",
				net.JoinHostPort(cluster.GetApp("es-test-server-v1").GetIP(), "8080"),
			)),
	).To(Succeed())

	Expect(cluster.GetKumactlOptions().
		KumactlApplyFromString(
			ExternalServiceV2(
				"default",
				net.JoinHostPort(cluster.GetApp("es-test-server-v2").GetIP(), "8080"),
			)),
	).To(Succeed())

	E2EDeferCleanup(func() {
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})
})

func StandaloneUniversal() {
	E2EAfterEach(func() {
		// remove all FaultInjections
		items, err := cluster.GetKumactlOptions().KumactlList("fault-injections", "default")
		Expect(err).ToNot(HaveOccurred())
		for _, item := range items {
			err := cluster.GetKumactlOptions().KumactlDelete("fault-injection", item, "default")
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
       kuma.io/service: demo-client
destinations:
   - match:
       kuma.io/service: external-service
       kuma.io/protocol: http
       version: v2
conf:
   abort:
     httpStatus: 401
     percentage: 100`)(cluster)).To(Succeed())

		Eventually(func() bool {
			stdout, _, err := cluster.Exec("", "", "demo-client",
				"curl", "-v", "-m", "8", "external-service.mesh")
			if err != nil {
				return false
			}
			return strings.Contains(stdout, "401 Unauthorized")
		}, "30s", "1s").Should(BeTrue())
	})
}
