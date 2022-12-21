package meshhealthcheck

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
)

func API() {
	meshName := "meshhealthcheck-api"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(
			DeleteMeshResources(env.Cluster, meshName, v1alpha1.MeshHealthCheckResourceTypeDescriptor),
		).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should create MeshHealthCheck policy", func() {
		// given no MeshHealthChecks
		mtps, err := env.Cluster.GetKumactlOptions().KumactlList("meshhealthchecks", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mtps).To(HaveLen(0))

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
`, Config.KumaNamespace))(env.Cluster)).To(Succeed())

		// then
		mtps, err = env.Cluster.GetKumactlOptions().KumactlList("meshhealthchecks", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mtps).To(HaveLen(1))
		Expect(mtps[0]).To(Equal(fmt.Sprintf("mhc1.%s", Config.KumaNamespace)))
	})

	It("should deny creating policy in the non-system namespace", func() {
		// given no MeshHealthChecks
		mtps, err := env.Cluster.GetKumactlOptions().KumactlList("meshhealthchecks", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mtps).To(HaveLen(0))

		// when
		err = k8s.KubectlApplyFromStringE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(), `
apiVersion: kuma.io/v1alpha1
kind: MeshHealthCheck
metadata:
  name: mhc1
  namespace: default
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
`)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("policy can only be created in the system namespace:%s", Config.KumaNamespace)))
	})
}
