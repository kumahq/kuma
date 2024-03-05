package meshfaultinjection

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func API() {
	meshName := "meshfaultinjection-api"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(kubernetes.Cluster, meshName, v1alpha1.MeshFaultInjectionResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should create MeshFaultInjection policy", func() {
		// given no MeshRateLimit
		mrls, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("meshfaultinjections", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mrls).To(BeEmpty())

		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshFaultInjection
metadata:
  name: mesh-fault-injection
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: backend
  from:
    - targetRef:
        kind: MeshServiceSubset
        name: frontend
        tags:
          kuma.io/zone: us-east
      default:
          http:
            - abort:
                httpStatus: 500
                percentage: "3.3"
              delay:
                value: 5s
                percentage: 3
              responseBandwidth:
                limit: 10mbps
                percentage: 1
            - delay:
                value: 11s
                percentage: "2.1"
    - targetRef:
        kind: MeshService
        name: test-server
      default:
          http:
            - abort:
                httpStatus: 500
                percentage: 3
            - delay:
                value: 5s
                percentage: "3.2"
            - responseBandwidth:
                limit: 10mbps
                percentage: 1
`, Config.KumaNamespace, meshName))(kubernetes.Cluster)).To(Succeed())

		// then
		mrls, err = kubernetes.Cluster.GetKumactlOptions().KumactlList("meshfaultinjections", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mrls).To(HaveLen(1))
		Expect(mrls[0]).To(Equal(fmt.Sprintf("mesh-fault-injection.%s", Config.KumaNamespace)))
	})

	It("should deny creating policy in the non-system namespace", func() {
		// given no MeshRateLimit
		mrls, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("meshfaultinjections", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mrls).To(BeEmpty())

		// when
		err = k8s.KubectlApplyFromStringE(
			kubernetes.Cluster.GetTesting(),
			kubernetes.Cluster.GetKubectlOptions(), fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshFaultInjection
metadata:
  name: mesh-fault-injection
  namespace: default
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: backend
  from:
    - targetRef:
        kind: MeshServiceSubset
        name: frontend
        tags:
          kuma.io/zone: us-east
      default:
          http:
            - abort:
                httpStatus: 500
                percentage: 3
              delay:
                value: 5s
                percentage: 3
              responseBandwidth:
                limit: 10mbps
                percentage: 1
            - delay:
                value: 11s
                percentage: 2
    - targetRef:
        kind: MeshService
        name: test-server
      default:
          http:
            - abort:
                httpStatus: 500
                percentage: 3
            - delay:
                value: 5s
                percentage: 3
            - responseBandwidth:
                limit: 10mbps
                percentage: 1
`, meshName))

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("policy can only be created in the system namespace:%s", Config.KumaNamespace)))
	})
}
