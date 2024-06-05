package meshservice

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func MeshService() {
	meshName := "meshservice"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugUniversal(multizone.UniZone2, meshName)
		DebugKube(multizone.KubeZone1, meshName)
		DebugKube(multizone.KubeZone2, meshName)
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should sync MeshService to global with VIP status", func() {
		ms := `
type: MeshService
name: backend
mesh: meshservice
labels:
  kuma.io/origin: zone
spec:
  selector:
    dataplaneTags:
      x: aaa
`

		// when
		Expect(multizone.UniZone1.Install(YamlUniversal(ms))).To(Succeed())

		// then VIP is assigned
		Eventually(func(g Gomega) {
			out, err := multizone.UniZone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshservices", "-m", meshName, "-oyaml")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(ContainSubstring("ip: 241.0.0."))
		}, "30s", "1s").Should(Succeed())

		// and MeshService is synced to global with status
		Eventually(func(g Gomega) {
			out, err := multizone.Global.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshservices", "-m", meshName, "-oyaml")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(ContainSubstring("kuma.io/display-name: backend"))
			g.Expect(out).To(ContainSubstring("ip: 241.0.0."))
		}, "30s", "1s").Should(Succeed())
	})
}
