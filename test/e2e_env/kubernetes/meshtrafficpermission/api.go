package meshtrafficpermission

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
)

func API() {
	meshName := "meshtrafficpermission-api"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should create MeshTrafficPermission policy", func() {
		// given no MeshTrafficPermissions
		mtps, err := env.Cluster.GetKumactlOptions().KumactlList("meshtrafficpermissions", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mtps).To(HaveLen(0))

		// when
		Expect(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: mtp1
  namespace: kuma-system
  labels:
    kuma.io/mesh: meshtrafficpermission-api
spec:
  targetRef:
    kind: MeshService
    name: backend
  from:
    - targetRef:
        kind: Mesh
      default:
        action: ALLOW
    - targetRef:
        kind: MeshSubset
        tags:
          kuma.io/zone: us-east
      default:
        action: DENY_WITH_SHADOW_ALLOW
    - targetRef:
        kind: MeshService
        name: backend
      default:
        action: ALLOW_WITH_SHADOW_DENY
    - targetRef:
        kind: MeshServiceSubset
        name: backend
        tags:
          version: v1
      default:
        action: DENY
`)(env.Cluster)).To(Succeed())

		// then
		mtps, err = env.Cluster.GetKumactlOptions().KumactlList("meshtrafficpermissions", meshName)
		Expect(err).ToNot(HaveOccurred())
		Expect(mtps).To(HaveLen(1))
		Expect(mtps[0]).To(Equal("mtp1.kuma-system"))
	})
}
