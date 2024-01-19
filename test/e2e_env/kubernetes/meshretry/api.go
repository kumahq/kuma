package meshretry

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func API() {
	meshName := "meshretry-api"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Delete the default meshretry policy
		Expect(DeleteMeshPolicyOrError(
			kubernetes.Cluster,
			v1alpha1.MeshRetryResourceTypeDescriptor,
			fmt.Sprintf("mesh-retry-all-%s", meshName),
		)).To(Succeed())
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

	It("should deny creating policy in the non-system namespace", func() {
		// given no MeshRetry
		mrls, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("meshretries", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mrls).To(BeEmpty())

		// when
		err = k8s.KubectlApplyFromStringE(
			kubernetes.Cluster.GetTesting(),
			kubernetes.Cluster.GetKubectlOptions(), fmt.Sprintf(`
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
