package meshtrafficpermission

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	. "github.com/kumahq/kuma/test/framework"
)

func MeshTrafficPermission() {
	const meshName = "mtp-test"
	const namespace = "mtp-test"

	var clientPodName string

	BeforeAll(func() {
		// Global
		err := env.Global.Install(MTLSMeshUniversal(meshName))
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(meshName, env.Zones())).To(Succeed())
		// remove default traffic permission
		Expect(env.Global.GetKumactlOptions().KumactlDelete("traffic-permission", "allow-all-"+meshName, meshName)).To(Succeed())

		// Universal Zone 1
		err = env.UniZone1.Install(TestServerUniversal("test-server", meshName,
			WithArgs([]string{"echo", "--instance", "echo"}),
		))
		Expect(err).ToNot(HaveOccurred())

		// Kubernetes Zone 1
		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(meshName, namespace)).
			Setup(env.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		clientPodName, err = PodNameOfApp(env.KubeZone1, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())
	})

	BeforeEach(func() {
		Expect(DeleteMeshResources(env.Global, meshName, policies_api.MeshTrafficPermissionResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Global.DeleteMesh(meshName)).To(Succeed())
	})

	trafficAllowed := func() {
		Eventually(func(g Gomega) {
			_, _, err := env.KubeZone1.Exec(namespace, clientPodName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
	}

	trafficBlocked := func() {
		Eventually(func() error {
			_, _, err := env.KubeZone1.Exec(namespace, clientPodName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "test-server.mesh")
			return err
		}).Should(HaveOccurred())
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
       action: ALLOW
`
		err := YamlUniversal(yaml)(env.Global)
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
       action: ALLOW
`
		err := YamlUniversal(yaml)(env.Global)
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
       action: ALLOW
`
		err := YamlUniversal(yaml)(env.Global)
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
       action: ALLOW
`
		err := YamlUniversal(yaml)(env.Global)
		Expect(err).ToNot(HaveOccurred())

		// and when Kubernetes pod is labeled
		err = k8s.RunKubectlE(env.KubeZone1.GetTesting(), env.KubeZone1.GetKubectlOptions(namespace), "label", "pod", clientPodName, "newtag=client")
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})
}
