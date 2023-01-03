package trafficpermission

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	. "github.com/kumahq/kuma/test/framework"
)

func TrafficPermission() {
	const meshName = "tp-test"
	const namespace = "tp-test"

	var clientPodName string

	BeforeAll(func() {
		// Global
		err := env.Global.Install(MTLSMeshUniversal(meshName))
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(meshName, env.Zones())).To(Succeed())

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
		Expect(DeleteMeshResources(env.Global, meshName, core_mesh.TrafficPermissionResourceTypeDescriptor)).To(Succeed())
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
		err := YamlUniversal(yaml)(env.Global)
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})

	It("should allow the traffic with kuma.io/zone", func() {
		// given
		trafficBlocked()

		// when
		yaml := `
type: TrafficPermission
mesh: tp-test
name: example-on-zone
sources:
 - match:
     kuma.io/zone: kuma-1-zone
destinations:
 - match:
     kuma.io/zone: kuma-4
`
		err := YamlUniversal(yaml)(env.Global)
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
		err := YamlUniversal(yaml)(env.Global)
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
		err := YamlUniversal(yaml)(env.Global)
		Expect(err).ToNot(HaveOccurred())

		// and when Kubernetes pod is labeled
		err = k8s.RunKubectlE(env.KubeZone1.GetTesting(), env.KubeZone1.GetKubectlOptions(namespace), "label", "pod", clientPodName, "newtag=client")
		Expect(err).ToNot(HaveOccurred())

		// then
		trafficAllowed()
	})
}
