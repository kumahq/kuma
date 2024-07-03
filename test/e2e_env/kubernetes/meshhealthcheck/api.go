package meshhealthcheck

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func API() {
	meshName := "meshhealthcheck-api"

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
		Expect(
			DeleteMeshResources(kubernetes.Cluster, meshName, v1alpha1.MeshHealthCheckResourceTypeDescriptor),
		).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should create MeshHealthCheck policy", func() {
		// given no MeshHealthChecks
		mtps, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("meshhealthchecks", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mtps).To(BeEmpty())

		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHealthCheck
metadata:
  name: mhc1
  namespace: %s
  labels:
    kuma.io/mesh: meshhealthcheck-api
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        interval: 10s
        timeout: 2s
        unhealthyThreshold: 3
        healthyThreshold: 1
        failTrafficOnPanic: true
        noTrafficInterval: 1s
        healthyPanicThreshold: 0
        reuseConnection: true
        http:
          path: /
          expectedStatuses: [200]
`, Config.KumaNamespace))(kubernetes.Cluster)).To(Succeed())

		// then
		mtps, err = kubernetes.Cluster.GetKumactlOptions().KumactlList("meshhealthchecks", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mtps).To(HaveLen(1))
		Expect(mtps[0]).To(Equal(fmt.Sprintf("mhc1.%s", Config.KumaNamespace)))
	})
}
