package trafficpermission

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func TrafficPermission() {
	const meshName = "tp-test"
	const namespace = "tp-test"

	var clientPodName string

	BeforeAll(func() {
		// Global
		err := NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(TrafficRouteUniversal(meshName)).
			Setup(multizone.Global)
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		// Universal Zone 1
		NewClusterSetup().
			Install(TestServerUniversal("test-server", meshName,
				WithArgs([]string{"echo", "--instance", "echo"}),
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
	})

	BeforeEach(func() {
		Expect(DeleteMeshResources(multizone.Global, meshName, core_mesh.TrafficPermissionResourceTypeDescriptor)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
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
			g.Expect(response.ResponseCode).To(Equal(403))
		}).Should(Succeed())
	}

	It("should allow the traffic with default traffic permission", func() {
		// given
		trafficBlocked()

		// when
		yaml := `
type: TrafficPermission
mesh: tp-test
name: allow-all-tp-test
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: '*'
`
		err := YamlUniversal(yaml)(multizone.Global)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should allow the traffic with kuma.io/zone", func() {
		// given
		trafficBlocked()

		// when
		yaml := fmt.Sprintf(`
type: TrafficPermission
mesh: tp-test
name: example-on-zone
sources:
 - match:
     kuma.io/zone: %s 
destinations:
 - match:
     kuma.io/zone: %s 
`, multizone.KubeZone1.ZoneName(), multizone.UniZone1.ZoneName())
		err := YamlUniversal(yaml)(multizone.Global)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should allow the traffic with k8s.kuma.io/namespace", func() {
		// given
		trafficBlocked()

		// when
		yaml := `
type: TrafficPermission
mesh: tp-test
name: example-on-namespace
sources:
 - match:
     k8s.kuma.io/namespace: tp-test
destinations:
 - match:
     kuma.io/service: test-server
`
		err := YamlUniversal(yaml)(multizone.Global)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should allow the traffic with tags added dynamically on Kubernetes", func() {
		// given
		trafficBlocked()

		// when
		yaml := `
type: TrafficPermission
mesh: tp-test
name: example-on-new-tag
sources:
 - match:
     newtag: client
destinations:
 - match:
     kuma.io/service: test-server
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
