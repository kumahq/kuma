package externalservices

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
)

func zoneExternalService(mesh string, ip string, name string, zone string) string {
	return fmt.Sprintf(`
type: ExternalService
mesh: "%s"
name: "%s"
tags:
  kuma.io/service: "%s"
  kuma.io/protocol: http
  kuma.io/zone: "%s"
networking:
  address: "%s"
`, mesh, name, name, zone, net.JoinHostPort(ip, "8080"))
}

const defaultMesh = "default"

var (
	global, zone1 Cluster
	zone4         *UniversalCluster
)

func ExternalServicesOnMultizoneHybridWithLocalityAwareLb() {
	BeforeAll(func() {
		// Global
		global = NewUniversalCluster(NewTestingT(), Kuma5, Silent)

		Expect(NewClusterSetup().
			Install(Kuma(config_core.Global)).
			Install(ResourceUniversal(samples.MeshMTLSBuilder().
				WithName(defaultMesh).
				WithEgressRoutingEnabled().
				WithoutPassthrough().Build())).
			Install(MeshTrafficPermissionAllowAllUniversal(defaultMesh)).
			Setup(global)).To(Succeed())

		globalCP := global.GetKuma()

		// K8s Cluster 1
		group := errgroup.Group{}
		zone1 = NewK8sCluster(NewTestingT(), Kuma1, Silent)
		group.Go(func() error {
			err := NewClusterSetup().
				Install(Kuma(config_core.Zone,
					WithIngress(),
					WithIngressEnvoyAdminTunnel(),
					WithEgress(),
					WithEgressEnvoyAdminTunnel(),
					WithGlobalAddress(globalCP.GetKDSServerAddress()),
				)).
				Install(NamespaceWithSidecarInjection(TestNamespace)).
				Setup(zone1)
			return errors.Wrap(err, zone1.Name())
		})

		// Universal Cluster 4
		zone4 = NewUniversalCluster(NewTestingT(), Kuma4, Silent)

		group.Go(func() error {
			err := NewClusterSetup().
				Install(Kuma(config_core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
				Install(DemoClientUniversal(
					"zone4-demo-client",
					defaultMesh,
					WithTransparentProxy(true),
				)).
				Install(IngressUniversal(globalCP.GenerateZoneIngressToken)).
				Install(EgressUniversal(globalCP.GenerateZoneEgressToken)).
				Install(TestServerExternalServiceUniversal("external-service-in-zone1", 8080, false)).
				Setup(zone4)
			return errors.Wrap(err, zone4.Name())
		})

		Expect(group.Wait()).To(Succeed())

		Expect(NewClusterSetup().
			Install(YamlUniversal(zoneExternalService(defaultMesh, zone4.GetApp("external-service-in-zone1").GetIP(), "external-service-in-zone1", "kuma-1"))).
			Setup(global),
		).To(Succeed())
	})

	AfterAll(func() {
		Expect(zone1.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zone1.DeleteKuma()).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())
		Expect(zone4.DismissCluster()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	DescribeTable("should fail request when zone proxy is down",
		func(fn func(c *K8sCluster) func() error) {
			k8sCluster := zone1.(*K8sCluster)
			Expect(k8sCluster.StartZoneEgress()).To(Succeed())
			Expect(k8sCluster.StartZoneIngress()).To(Succeed())
			// when
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					zone4, "zone4-demo-client", "external-service-in-zone1.mesh")
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			// when ingress is down
			Expect(fn(k8sCluster)()).To(Succeed())

			// then service is unreachable
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					zone4, "zone4-demo-client", "external-service-in-zone1.mesh")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(503))
			}, "30s").ShouldNot(HaveOccurred())
		},
		Entry("egress", func(c *K8sCluster) func() error { return c.StopZoneEgress }),
		Entry("ingress", func(c *K8sCluster) func() error { return c.StopZoneIngress }),
	)
}
