package meshtrafficpermission

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func MeshTrafficPermission() {
	const meshName = "mtp-test"
	const namespace = "mtp-test"

	var clientPodName string

	BeforeAll(func() {
		// Global
		err := multizone.Global.Install(MTLSMeshUniversal(meshName))
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())
		// remove default traffic permission
		Expect(multizone.Global.GetKumactlOptions().KumactlDelete("traffic-permission", "allow-all-"+meshName, meshName)).To(Succeed())

		// Universal Zone 1
		err = multizone.UniZone1.Install(TestServerUniversal("test-server", meshName,
			WithArgs([]string{"echo", "--instance", "echo"}),
		))
		Expect(err).ToNot(HaveOccurred())

		// Kubernetes Zone 1
		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName))).
			Setup(multizone.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		clientPodName, err = PodNameOfApp(multizone.KubeZone1, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())
	})

	BeforeEach(func() {
		Expect(DeleteMeshResources(multizone.Global, meshName, policies_api.MeshTrafficPermissionResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	trafficAllowed := func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client", "test-server.mesh",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
	}

	trafficBlocked := func() {
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				multizone.KubeZone1, "demo-client", "test-server.mesh",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(503))
		}).Should(Succeed())
	}

	It("should allow the traffic with allow-all meshtrafficpermission", func() {
		// given no mesh traffic permissions
		trafficBlocked()

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
		trafficAllowed()
	})

	It("should allow the traffic with kuma.io/zone", func() {
		// given no mesh traffic permissions
		trafficBlocked()

		// when mesh traffic permission with MeshService
		yaml := `
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
         kuma.io/zone: kuma-1-zone
     default:
       action: Allow
`
		err := YamlUniversal(yaml)(multizone.Global)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should allow the traffic with k8s.kuma.io/namespace", func() {
		// given no mesh traffic permissions
		trafficBlocked()

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
		trafficAllowed()
	})

	It("should allow the traffic with tags added dynamically on Kubernetes", func() {
		// given no mesh traffic permissions
		trafficBlocked()

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
		trafficAllowed()
	})
}
