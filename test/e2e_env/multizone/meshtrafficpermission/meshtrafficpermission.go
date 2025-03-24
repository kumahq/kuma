package meshtrafficpermission

import (
	"fmt"
	"net"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func externalService(mesh string, ip string) InstallFunc {
	return YamlUniversal(fmt.Sprintf(`
type: ExternalService
mesh: "%s"
name: es-1
tags:
  kuma.io/service: external-service
  kuma.io/protocol: http
networking:
  address: "%s"
`, mesh, net.JoinHostPort(ip, "80")))
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

	var clientPodName string

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
				TestServerUniversal("test-server", meshName,
					WithArgs([]string{"echo", "--instance", "echo"}),
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

		clientPodName, err = PodNameOfApp(multizone.KubeZone1, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())

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
		// given no mesh traffic permissions
		trafficBlocked("test-server.mesh")

		// when mesh traffic permission with MeshService
		yaml := `
type: MeshTrafficPermission
name: mtp-1
mesh: mtp-test
spec:
 targetRef:
   kind: Mesh
 from:
   - targetRef:
       kind: Mesh
     default:
       action: Allow
`
		err := YamlUniversal(yaml)(multizone.Global)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed("test-server.mesh")
	})

	It("should allow the traffic with kuma.io/zone", func() {
		// given no mesh traffic permissions
		trafficBlocked("test-server.mesh")

		// when mesh traffic permission with MeshService
		yaml := fmt.Sprintf(`
type: MeshTrafficPermission
name: mtp-2
mesh: mtp-test
spec:
 targetRef:
   kind: MeshService
   name: test-server
 from:
   - targetRef:
       kind: MeshSubset
       tags:
         kuma.io/zone: %s 
     default:
       action: Allow
`, multizone.KubeZone1.ZoneName())
		err := YamlUniversal(yaml)(multizone.Global)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed("test-server.mesh")
	})

	It("should allow the traffic with k8s.kuma.io/namespace", func() {
		// given no mesh traffic permissions
		trafficBlocked("test-server.mesh")

		// when mesh traffic permission with MeshSubset
		yaml := `
type: MeshTrafficPermission
name: mtp-3
mesh: mtp-test
spec:
 targetRef:
   kind: MeshService
   name: test-server
 from:
   - targetRef:
       kind: MeshSubset
       tags:
         k8s.kuma.io/namespace: mtp-test
     default:
       action: Allow
`
		err := YamlUniversal(yaml)(multizone.Global)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed("test-server.mesh")
	})

	It("should allow the traffic with tags added dynamically on Kubernetes", func() {
		// given no mesh traffic permissions
		trafficBlocked("test-server.mesh")

		// when mesh traffic permission with MeshSubset
		yaml := `
type: MeshTrafficPermission
name: mtp-4
mesh: mtp-test
spec:
 targetRef:
   kind: MeshService
   name: test-server
 from:
   - targetRef:
       kind: MeshSubset
       tags:
         newtag: client
     default:
       action: Allow
`
		err := YamlUniversal(yaml)(multizone.Global)
		Expect(err).ToNot(HaveOccurred())

		// and when Kubernetes pod is labeled
		err = k8s.RunKubectlE(multizone.KubeZone1.GetTesting(), multizone.KubeZone1.GetKubectlOptions(namespace), "label", "pod", clientPodName, "newtag=client")
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed("test-server.mesh")
	})

	It("should allow the traffic to the external service through the egress", func() {
		// given no mesh traffic permissions
		trafficBlocked("external-service.mesh")

		// when mesh traffic permission with MeshSubset
		yaml := `
type: MeshTrafficPermission
name: mtp-5
mesh: mtp-test
spec:
 targetRef:
   kind: MeshService
   name: external-service
 from:
   - targetRef:
       kind: Mesh
     default:
       action: Allow
`
		err := YamlUniversal(yaml)(multizone.Global)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed("external-service.mesh")
	})
}
