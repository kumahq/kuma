package meshservice

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func MeshService() {
	meshName := "mesh-service"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(TestServerUniversal(
				"first-test-server",
				meshName,
				WithArgs([]string{"echo"}),
				WithServiceName("first-test-server"),
				WithAdditionalTags(map[string]string{
					"app": "test-server",
				}))).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should be able to create and use MeshService with HostnameGenerator", func() {
		generator := `
type: HostnameGenerator
mesh: mesh-service
name: basic
labels:
  kuma.io/origin: zone
spec:
  template: "{{ .Name }}.mesh"
  selector:
    meshService:
      matchLabels:
        app: backend
`
		service := fmt.Sprintf(`
type: MeshService
name: backend
mesh: "%s"
labels:
  app: backend
  kuma.io/origin: zone
spec:
  ports:
  - port: 80 
    targetPort: 80
    protocol: http
  selector:
    dataplaneTags:
      app: test-server
`, meshName)
		Expect(universal.Cluster.Install(YamlUniversal(generator))).To(Succeed())
		Expect(universal.Cluster.Install(YamlUniversal(service))).To(Succeed())

		Eventually(func(g Gomega) {
			// when
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "backend.mesh",
			)
			// then
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "500ms", MustPassRepeatedly(10)).Should(Succeed())
	})
}
