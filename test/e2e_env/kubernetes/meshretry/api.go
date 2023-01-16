package meshretry

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
)

func API() {
	meshName := "meshretry-api"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(env.Cluster, meshName, v1alpha1.MeshRetryResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should create MeshRetry policy", func() {
		// given no MeshRetry
		mrls, err := env.Cluster.GetKumactlOptions().KumactlList("meshretrys", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mrls).To(HaveLen(0))

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
`, Config.KumaNamespace, meshName))(env.Cluster)).To(Succeed())

		// then
		mrls, err = env.Cluster.GetKumactlOptions().KumactlList("meshretrys", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mrls).To(HaveLen(1))
		Expect(mrls[0]).To(Equal(fmt.Sprintf("mesh-retry.%s", Config.KumaNamespace)))
	})

	It("should deny creating policy in the non-system namespace", func() {
		// given no MeshRetry
		mrls, err := env.Cluster.GetKumactlOptions().KumactlList("meshretrys", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mrls).To(HaveLen(0))

		// when
		err = k8s.KubectlApplyFromStringE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(), fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshRetry
metadata:
  name: mesh-retry
  namespace: default
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
`, meshName))

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("policy can only be created in the system namespace:%s", Config.KumaNamespace)))
	})
}
