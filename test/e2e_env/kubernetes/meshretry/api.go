package meshretry

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

func API() {
	meshName := "meshretry-api"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(Yaml(builders.Mesh().
				WithName(meshName).
				WithoutInitialPolicies())).
			Setup(kubernetes.Cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, Config.KumaNamespace)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(kubernetes.Cluster, meshName, v1alpha1.MeshRetryResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should create MeshRetry policy", func() {
		// given no MeshRetry
		mrls, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("meshretries", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mrls).To(BeEmpty())

		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshRetry
metadata:
  name: mesh-retry
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: backend
      default:
        tcp:
          maxConnectAttempt: 5
        http:
          numRetries: 10
`, Config.KumaNamespace, meshName))(kubernetes.Cluster)).To(Succeed())

		// then
		mrls, err = kubernetes.Cluster.GetKumactlOptions().KumactlList("meshretries", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mrls).To(HaveLen(1))
		Expect(mrls[0]).To(Equal(fmt.Sprintf("mesh-retry.%s", Config.KumaNamespace)))
	})
}
