package meshtrafficpermission

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	policies_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v3/test/framework/envs/multizone"
)

func externalService(mesh string, ip string) InstallFunc {
	return YamlUniversal(fmt.Sprintf(`
type: MeshExternalService
name: external-service
mesh: "%s"
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: "%s"
      port: 80
`, mesh, ip))
}

func mtlsAndEgressMeshUniversal(name string) InstallFunc {
	mesh := fmt.Sprintf(`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
    - name: ca-1
      type: builtin
routing:
  zoneEgress: true
`, name)
	return YamlUniversal(mesh)
}

func MeshTrafficPermission() {
	const meshName = "mtp-test"
	const namespace = "mtp-test"

	BeforeAll(func() {
		// Global
		err := NewClusterSetup().
			Install(mtlsAndEgressMeshUniversal(meshName)).
			Setup(multizone.Global)
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		// Universal Zone 1
		NewClusterSetup().
			Install(Parallel(
				TestServerUniversal(
					"test-server", meshName,
					WithArgs([]string{"echo", "--instance", "echo"}),
					WithLabels(map[string]string{"kuma.io/service": "test-server"}),
				),
				TestServerExternalServiceUniversal("external-service", 80, false, WithDockerContainerName("kuma-es-4_external-service-mtp-test")),
			)).
			SetupInGroup(multizone.UniZone1, &group)

		// Kubernetes Zone 1
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName))).
			SetupInGroup(multizone.KubeZone1, &group)
		Expect(group.Wait()).To(Succeed())

		esIp := multizone.UniZone1.GetApp("external-service").GetIP()
		Expect(multizone.Global.Install(externalService(meshName, esIp))).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
	})

	BeforeEach(func() {
		Expect(DeleteMeshResources(multizone.Global, meshName, policies_api.MeshTrafficPermissionResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone1.DeleteApp("external-service")).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	trafficAllowed := func(url string) {
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client", url,
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
	}

	trafficBlocked := func(url string) {
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				multizone.KubeZone1, "demo-client", url,
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(403))
		}).Should(Succeed())
	}

	It("should allow the traffic with allow-all meshtrafficpermission", func() {
		serverHostname := fmt.Sprintf("test-server.svc.%s.mesh.local", multizone.UniZone1.ZoneName())

		trafficBlocked(serverHostname)

		yaml := `
type: MeshTrafficPermission
name: mtp-1
mesh: mtp-test
spec:
 targetRef:
   kind: Mesh
 rules:
   - default:
       allow:
         - spiffeID:
             type: Prefix
             value: spiffe://mtp-test
`
		err := YamlUniversal(yaml)(multizone.Global)
		Expect(err).ToNot(HaveOccurred())

		trafficAllowed(serverHostname)
	})

	It("should allow the traffic to the external service through the egress", func() {
		Skip("MeshTrafficPermission cannot gate a MeshExternalService without Zone Proxy + MeshIdentity (SNI rules); tracked in https://github.com/kumahq/kuma/issues/17160")
		trafficAllowed("external-service.extsvc.mesh.local")
	})
}
