package inspect

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
)

func Inspect() {
	meshName := "inspect"

	BeforeAll(func() {
		E2EDeferCleanup(func() {
			Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
			Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(DemoClientUniversal(AppModeDemoClient, meshName)).
			Setup(env.Cluster)
		Expect(err).To(BeNil())
	})

	It("should return envoy config_dump", func() {
		Eventually(func(g Gomega) {
			output, err := env.Cluster.GetKumactlOptions().
				RunKumactlAndGetOutput("inspect", "dataplane", AppModeDemoClient, "--config-dump",
					"--mesh", meshName)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(output).To(ContainSubstring(`"name": "kuma:envoy:admin"`))
			g.Expect(output).To(ContainSubstring(`"name": "outbound:127.0.0.1:4000"`))
			g.Expect(output).To(ContainSubstring(`"name": "outbound:127.0.0.1:4001"`))
			g.Expect(output).To(ContainSubstring(`"name": "outbound:127.0.0.1:5000"`))
		}, "30s", "1s").Should(Succeed())
	})
}
