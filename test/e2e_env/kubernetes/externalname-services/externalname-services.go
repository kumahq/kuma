package externalname_services

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func ExternalNameServices() {
	meshName := "externalname-services"
	namespace := "externalname-services"

	externalNameService := fmt.Sprintf(`
apiVersion: v1
kind: Service
metadata:
  name: externalname-service
  namespace: %s
spec:
  type: ExternalName
  externalName: foo.bar
  ports: 
  - appProtocol: tcp
    port: 3000
    protocol: TCP
    targetPort: 3000
  selector:
    app: test-server
`, namespace)

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
			)).
			Setup(kubernetes.Cluster),
		).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should ignore ExternalName Service", func() {
		// when
		Expect(kubernetes.Cluster.Install(YamlK8s(externalNameService))).To(Succeed())

		// then
		Consistently(kubernetes.Cluster.GetKumaCPLogs, "10s", "1s").
			ShouldNot(ContainSubstring("could not parse hostname entry"))
	})
}
