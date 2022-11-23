package mestimeout

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
)

func API() {
	meshName := "meshtimeout-api"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(
			k8s.RunKubectlE(env.Cluster.GetTesting(), env.Cluster.GetKubectlOptions(), "delete", "meshtimeouts", "-A", "--all"),
		).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should create MeshTrafficPermission policy", func() {
		// given no MeshTimeout
		mts, err := env.Cluster.GetKumactlOptions().KumactlList("meshtimeouts", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mts).To(HaveLen(0))

		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1
  namespace: %s
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: kuma-demo
      default:
        idleTimeout: 20s
        http:
          requestTimeout: 20s
          maxStreamDuration: 20s
  from:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 10s
        http:
          requestTimeout: 10s
          maxStreamDuration: 10s
`, Config.KumaNamespace))(env.Cluster)).To(Succeed())

		// then
		mts, err = env.Cluster.GetKumactlOptions().KumactlList("meshtimeouts", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mts).To(HaveLen(1))
		Expect(mts[0]).To(Equal(fmt.Sprintf("mt1.%s", Config.KumaNamespace)))
	})
}
