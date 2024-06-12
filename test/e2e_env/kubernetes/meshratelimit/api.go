package meshratelimit

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func API() {
	meshName := "meshratelimit-api"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, Config.KumaNamespace)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(kubernetes.Cluster, meshName, v1alpha1.MeshRateLimitResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should create MeshRateLimit policy", func() {
		// given no MeshRateLimit
		mrls, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("meshratelimits", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mrls).To(BeEmpty())

		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshRateLimit
metadata:
  name: mesh-rate-limit
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: backend
  from:
    - targetRef:
        kind: Mesh
      default:
        local:
          http:
            requestRate:
              num: 1
              interval: 10s
            onRateLimit:
              status: 429
              headers:
                add:
                - name: "x-kuma-rate-limited"
                  value: "true"
    - targetRef:
        kind: Mesh
      default:
        local:
          http:
            requestRate:
              num: 1
              interval: 10s
            onRateLimit:
              status: 429
              headers:
                add:
                - name: "x-kuma-rate-limited"
                  value: "true"
          tcp:
            connectionRate:
              num: 100
              interval: 10s
`, Config.KumaNamespace, meshName))(kubernetes.Cluster)).To(Succeed())

		// then
		mrls, err = kubernetes.Cluster.GetKumactlOptions().KumactlList("meshratelimits", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mrls).To(HaveLen(1))
		Expect(mrls[0]).To(Equal(fmt.Sprintf("mesh-rate-limit.%s", Config.KumaNamespace)))
	})
}
