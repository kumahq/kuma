package meshservicelabelpropagation

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v2/test/framework"
)

var Cluster *framework.UniversalCluster

func LabelPropagation() {
	const meshName = "lp-mesh"
	const dpName = "lp-dp-1"
	const serviceTag = "lp-svc"

	BeforeAll(func() {
		Expect(framework.NewClusterSetup().
			Install(framework.MTLSMeshWithMeshServicesUniversal(meshName, "Exclusive")).
			Setup(Cluster)).To(Succeed())
		Expect(framework.WaitForMesh(meshName, []framework.Cluster{Cluster})).To(Succeed())
	})

	framework.AfterEachFailure(func() {
		framework.DebugUniversal(Cluster, meshName)
	})

	framework.E2EAfterAll(func() {
		_ = Cluster.GetKumactlOptions().RunKumactl("delete", "dataplane", dpName, "-m", meshName)
		Expect(Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("propagates non-reserved Dataplane tags and labels to auto-generated MeshService", func() {
		dp := `
type: Dataplane
mesh: lp-mesh
name: lp-dp-1
labels:
  color: blue
  kuma.io/owner: ignored
networking:
  address: 192.168.10.10
  inbound:
  - port: 80
    tags:
      kuma.io/service: lp-svc
      kuma.io/protocol: http
      team: payments
`
		Expect(framework.YamlUniversal(dp)(Cluster)).To(Succeed())

		Eventually(func(g Gomega) {
			labels, err := framework.GetMeshServiceLabels(Cluster, meshName, serviceTag)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(labels).To(HaveKeyWithValue("color", "blue"))
			g.Expect(labels).To(HaveKeyWithValue("team", "payments"))

			g.Expect(labels).ToNot(HaveKey("kuma.io/owner"))
			g.Expect(labels).ToNot(HaveKey("kuma.io/protocol"))

			g.Expect(labels).To(HaveKeyWithValue(metadata.KumaMeshLabel, meshName))
			g.Expect(labels).To(HaveKeyWithValue(mesh_proto.ManagedByLabel, "meshservice-generator"))
			g.Expect(labels).To(HaveKeyWithValue(mesh_proto.ResourceOriginLabel, string(mesh_proto.ZoneResourceOrigin)))
		}, "60s", "2s").Should(Succeed())
	})

	It("propagates updated tag value when Dataplane is re-applied", func() {
		// Change team payments→platform and add tier to cover both value-update and addition paths.
		dpUpdated := `
type: Dataplane
mesh: lp-mesh
name: lp-dp-1
labels:
  color: blue
networking:
  address: 192.168.10.10
  inbound:
  - port: 80
    tags:
      kuma.io/service: lp-svc
      kuma.io/protocol: http
      team: platform
      tier: backend
`
		Expect(framework.YamlUniversal(dpUpdated)(Cluster)).To(Succeed())

		Eventually(func(g Gomega) {
			labels, err := framework.GetMeshServiceLabels(Cluster, meshName, serviceTag)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(labels).To(HaveKeyWithValue("team", "platform"))
			g.Expect(labels).ToNot(HaveKeyWithValue("team", "payments"))
			g.Expect(labels).To(HaveKeyWithValue("tier", "backend"))
			g.Expect(labels).To(HaveKeyWithValue("color", "blue"))
		}, "60s", "2s").Should(Succeed())
	})
}
