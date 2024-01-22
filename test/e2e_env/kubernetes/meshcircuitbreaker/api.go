package meshcircuitbreaker

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func API() {
	meshName := "meshcircuitbreaker-api"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Delete the default meshcricuitbreaker policy
		Expect(DeleteMeshPolicyOrError(
			kubernetes.Cluster,
			v1alpha1.MeshCircuitBreakerResourceTypeDescriptor,
			fmt.Sprintf("mesh-circuit-breaker-all-%s", meshName),
		)).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(kubernetes.Cluster, meshName, v1alpha1.MeshCircuitBreakerResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should create MeshCircuitBreaker policy", func() {
		// given no MeshCircuitBreakers
		mcb, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("meshcircuitbreakers", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mcb).To(BeEmpty())

		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshCircuitBreaker
metadata:
  name: mcb-api-1
  namespace: %s
  labels:
    kuma.io/mesh: meshcircuitbreaker-api
spec:
  targetRef:
    kind: MeshService
    name: backend
  from:
    - targetRef:
        kind: Mesh
      default:
        connectionLimits:
          maxConnectionPools: 5
          maxConnections: 10
          maxPendingRequests: 2
          maxRetries: 1
          maxRequests: 32
  to:
    - targetRef:
        kind: MeshService
        name: frontend
      default:
        outlierDetection:
          disabled: false
          interval: 5s
          baseEjectionTime: 30s
          maxEjectionPercent: 20
          splitExternalAndLocalErrors: true
          detectors:
            totalFailures:
              consecutive: 10
            gatewayFailures:
              consecutive: 10
            localOriginFailures:
              consecutive: 10
            successRate:
              minimumHosts: 5
              requestVolume: 10
              standardDeviationFactor: 1900
            failurePercentage:
              requestVolume: 10
              minimumHosts: 5
              threshold: 85
`, Config.KumaNamespace))(kubernetes.Cluster)).To(Succeed())

		// then
		mcb, err = kubernetes.Cluster.GetKumactlOptions().KumactlList("meshcircuitbreakers", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mcb).To(HaveLen(1))
		Expect(mcb[0]).To(Equal(fmt.Sprintf("mcb-api-1.%s", Config.KumaNamespace)))
	})

	It("should deny creating policy in the non-system namespace", func() {
		// given no MeshCircuitBreakers
		mcb, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("meshcircuitbreakers", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mcb).To(BeEmpty())

		// when
		err = k8s.KubectlApplyFromStringE(
			kubernetes.Cluster.GetTesting(),
			kubernetes.Cluster.GetKubectlOptions(), `
apiVersion: kuma.io/v1alpha1
kind: MeshCircuitBreaker
metadata:
  name: mcb-api-invalid
  namespace: default
  labels:
    kuma.io/mesh: meshcircuitbreaker-api
spec:
  targetRef:
    kind: MeshService
    name: backend
  from:
    - targetRef:
        kind: Mesh
      default:
        connectionLimits:
          maxConnectionPools: 5
          maxConnections: 10
          maxPendingRequests: 2
          maxRetries: 1
          maxRequests: 32
  to:
    - targetRef:
        kind: MeshService
        name: frontend
      default:
        outlierDetection:
          disabled: false
          interval: 5s
          baseEjectionTime: 30s
          maxEjectionPercent: 20
          splitExternalAndLocalErrors: true
          detectors:
            totalFailures:
              consecutive: 10
            gatewayFailures:
              consecutive: 10
            localOriginFailures:
              consecutive: 10
            successRate:
              minimumHosts: 5
              requestVolume: 10
              standardDeviationFactor: 1900
            failurePercentage:
              requestVolume: 10
              minimumHosts: 5
              threshold: 85
`)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("policy can only be created in the system namespace:%s", Config.KumaNamespace)))
	})
}
