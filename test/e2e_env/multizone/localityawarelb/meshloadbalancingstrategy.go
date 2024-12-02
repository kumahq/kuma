package localityawarelb

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func LocalityAwarenessWithMeshLoadBalancingStrategy() {
	zoneExternalService := func(mesh, ip, name, service, zone string) string {
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
`, mesh, name, service, zone, net.JoinHostPort(ip, "8080"))
	}

	const mesh = "la-with-mesh-lb-strategy"

	BeforeAll(func() {
		// Global
		Expect(NewClusterSetup().
			Install(YamlUniversal(MeshMTLSOnAndZoneEgressAndNoPassthrough(mesh, "false"))).
			Install(MeshTrafficPermissionAllowAllUniversal(mesh)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(mesh, multizone.Zones())).To(Succeed())

		// Universal Zone 4
		group := errgroup.Group{}
		NewClusterSetup().
			Install(Parallel(
				DemoClientUniversal("demo-client", mesh, WithTransparentProxy(true)),
				InstallExternalService("es-1"),
				InstallExternalService("es-2"),
				TestServerUniversal("test-server-1", mesh,
					WithArgs([]string{"echo", "--instance", "ts-1"}),
				),
			)).
			SetupInGroup(multizone.UniZone1, &group)

		NewClusterSetup().
			Install(TestServerUniversal("test-server-2", mesh,
				WithArgs([]string{"echo", "--instance", "ts-2"}),
			)).
			SetupInGroup(multizone.UniZone2, &group)

		Expect(group.Wait()).To(Succeed())

		es1 := multizone.UniZone1.GetApp("es-1").GetIP()
		es2 := multizone.UniZone1.GetApp("es-2").GetIP()

		Expect(NewClusterSetup().
			Install(YamlUniversal(zoneExternalService(mesh, es1, "es-1", "es", "kuma-4"))).
			Install(YamlUniversal(zoneExternalService(mesh, es2, "es-2", "es", "kuma-1-zone"))).
			Setup(multizone.Global)).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, mesh)
		DebugUniversal(multizone.UniZone1, mesh)
		DebugUniversal(multizone.UniZone2, mesh)
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone1.DeleteMeshApps(mesh)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(mesh)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(mesh)).To(Succeed())
	})

	It("should load balance to the same ES when LA is enabled", func() {
		Expect(YamlUniversal(fmt.Sprintf(`
type: MeshLoadBalancingStrategy
name: mlbs-1
mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: demo-client
  to:
    - targetRef:
        kind: MeshService
        name: es
      default:
        localityAwarenes:
          disabled: false`, mesh))(multizone.Global)).To(Succeed())

		Eventually(func(g Gomega) {
			response, err := client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "es.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response).To(HaveLen(1))
		}, "30s", "500ms").Should(Succeed())
	})

	It("should load balance to the same internal service when LA is disabled", func() {
		Expect(YamlUniversal(fmt.Sprintf(`
type: MeshLoadBalancingStrategy
name: mlbs-1
mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: demo-client
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        localityAwarenes:
          disabled: false`, mesh))(multizone.Global)).To(Succeed())

		Eventually(func(g Gomega) {
			response, err := client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response).To(HaveLen(1))
		}, "30s", "500ms").Should(Succeed())
	})
}
