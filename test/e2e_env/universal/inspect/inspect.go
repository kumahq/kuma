package inspect

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Inspect() {
	meshName := "inspect"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(DemoClientUniversal(AppModeDemoClient, meshName)).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should return envoy config_dump", func() {
		Eventually(func(g Gomega) {
			output, err := universal.Cluster.GetKumactlOptions().
				RunKumactlAndGetOutput("inspect", "dataplane", AppModeDemoClient, "--type", "config-dump",
					"--mesh", meshName)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(output).To(ContainSubstring(`"name": "kuma:envoy:admin"`))
			g.Expect(output).To(ContainSubstring(`"name": "outbound:127.0.0.1:4000"`))
			g.Expect(output).To(ContainSubstring(`"name": "outbound:127.0.0.1:4001"`))
			g.Expect(output).To(ContainSubstring(`"name": "outbound:127.0.0.1:5000"`))
		}, "30s", "1s").Should(Succeed())
	})

	It("should return stats", func() {
		Eventually(func(g Gomega) {
			output, err := universal.Cluster.GetKumactlOptions().
				RunKumactlAndGetOutput("inspect", "dataplane", AppModeDemoClient, "--type", "stats",
					"--mesh", meshName)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(output).To(ContainSubstring(`server.live: 1`))
		}, "30s", "1s").Should(Succeed())
	})

	It("should return clusters", func() {
		Eventually(func(g Gomega) {
			output, err := universal.Cluster.GetKumactlOptions().
				RunKumactlAndGetOutput("inspect", "dataplane", AppModeDemoClient, "--type", "clusters",
					"--mesh", meshName)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(output).To(ContainSubstring(`kuma:envoy:admin::`))
		}, "30s", "1s").Should(Succeed())
	})
}
