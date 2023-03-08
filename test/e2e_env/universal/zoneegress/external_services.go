package zoneegress

import (
	"fmt"
	"net"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/envs/universal"
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

func ExternalServices() {
	meshName := "ze-external-services"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(YamlUniversal(meshMTLSOn(meshName))).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Install(ExternalServerUniversal("zef-test-server-v1")).
			Install(ExternalServerUniversal("zef-test-server-v2")).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(YamlUniversal(ExternalServiceV1(
				meshName,
				net.JoinHostPort(universal.Cluster.GetApp("zef-test-server-v1").GetIP(), "8080"),
			))).
			Install(YamlUniversal(ExternalServiceV2(
				meshName,
				net.JoinHostPort(universal.Cluster.GetApp("zef-test-server-v2").GetIP(), "8080"),
			))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	Context("Proxy", func() {
		filter := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			meshName,
			"external-service",
		)

		It("should access external service through zoneegress", func() {
			Eventually(func(g Gomega) {
				stat, err := universal.Cluster.GetZoneEgressEnvoyTunnel().GetStats(filter)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stat).To(stats.BeEqualZero())
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				stdout, _, err := client.CollectRawResponse(
					universal.Cluster, "demo-client", "external-service.mesh",
					client.WithVerbose(),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				stat, err := universal.Cluster.GetZoneEgressEnvoyTunnel().GetStats(filter)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stat).To(stats.BeGreaterThanZero())
			}).Should(Succeed())
		})
	})

	Context("Fault Injection", func() {
		AfterEach(func() {
			Expect(DeleteMeshResources(universal.Cluster, meshName, core_mesh.FaultInjectionResourceTypeDescriptor)).To(Succeed())
		})

		It("should inject faults for external service", func() {
			Expect(YamlUniversal(`
type: FaultInjection
mesh: ze-external-services
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
     percentage: 100`)(universal.Cluster)).To(Succeed())

			Eventually(func() bool {
				stdout, _, err := client.CollectRawResponse(
					universal.Cluster, "demo-client", "external-service.mesh",
					client.WithVerbose(),
					client.WithMaxTime(8),
				)
				if err != nil {
					return false
				}
				return strings.Contains(stdout, "401 Unauthorized")
			}, "30s", "1s").Should(BeTrue())
		})
	})

	Context("Rate Limit", func() {
		AfterEach(func() {
			Expect(DeleteMeshResources(universal.Cluster, meshName, core_mesh.RateLimitResourceTypeDescriptor)).To(Succeed())
		})

		It("should rate limit requests to external service", func() {
			specificRateLimitPolicy := `
type: RateLimit
mesh: ze-external-services
name: rate-limit-demo-client
sources:
- match:
    kuma.io/service: demo-client
destinations:
- match:
    kuma.io/service: external-service
conf:
  http:
    onRateLimit:
      status: 429
    requests: 1
    interval: 10s
`
			Expect(universal.Cluster.Install(YamlUniversal(specificRateLimitPolicy))).To(Succeed())

			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					universal.Cluster, "demo-client", "external-service.mesh",
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(429))
			}).Should(Succeed())
		})
	})
}
