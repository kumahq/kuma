package meshservice

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/kds/hash"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func Migration() {
	namespace := "msmigration"
	meshName := "msmigration"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Setup(multizone.Global)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName("demo-client"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
			)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
			)).
			Setup(multizone.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(TestServerUniversal("test-server", meshName)).
			Setup(multizone.UniZone1)
		Expect(err).ToNot(HaveOccurred())
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

	type meshServiceList struct {
		Total int `json:"total"`
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
	}

	unmarshal := func(out string) *meshServiceList {
		l := &meshServiceList{}
		Expect(yaml.Unmarshal([]byte(out), l)).To(Succeed())
		return l
	}

	hasMeshServices := func(names ...string) {
		Eventually(func(g Gomega) {
			// when
			out, err := multizone.Global.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshservices", "-m", meshName, "-o", "yaml")
			// then
			g.Expect(err).ToNot(HaveOccurred())
			response := unmarshal(out)
			g.Expect(response.Total).To(Equal(len(names)))
			var actualNames []string
			for _, name := range response.Items {
				actualNames = append(actualNames, name.Name)
			}
			g.Expect(actualNames).To(ConsistOf(names))
		}).Should(Succeed())
	}

	noMeshServices := func() {
		// call hasMeshServices with no names
		hasMeshServices()
	}

	It("should automatically create MeshServices when the mode is 'Exclusive'", func() {
		// given
		noMeshServices()

		// when enable 'mode: Exclusive'
		Expect(MTLSMeshWithMeshServicesUniversal(meshName, "Exclusive")(multizone.Global)).To(Succeed())

		// then
		hasMeshServices(
			hash.HashedName(meshName, "demo-client", Kuma1, namespace),
			hash.HashedName(meshName, "test-server", Kuma1, namespace),
			hash.HashedName(meshName, "test-server", Kuma4),
		)
	})

	It("should be possible to manually create MeshService on Universal", func() {
		// when
		Expect(YamlUniversal(fmt.Sprintf(`
type: MeshService
name: manually-created-ms
mesh: %s
labels:
  kuma.io/origin: zone
spec:
  selector:
    dataplaneTags:
      kuma.io/service: manually-created-ms
  ports:
  - port: 80
    targetPort: 80
    appProtocol: http
`, meshName))(multizone.UniZone1)).To(Succeed())

		// then
		hasMeshServices(
			hash.HashedName(meshName, "demo-client", Kuma1, namespace),
			hash.HashedName(meshName, "test-server", Kuma1, namespace),
			hash.HashedName(meshName, "test-server", Kuma4),
			hash.HashedName(meshName, "manually-created-ms", Kuma4),
		)
	})

	It("should delete automatically created MeshServices when the mode is 'Disabled'", func() {
		// when mode is 'Disabled'
		Expect(MTLSMeshWithMeshServicesUniversal(meshName, "Disabled")(multizone.Global)).To(Succeed())

		// then
		hasMeshServices(
			hash.HashedName(meshName, "manually-created-ms", Kuma4),
		)
	})
}
