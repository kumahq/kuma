package inspect

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	. "github.com/kumahq/kuma/test/framework"
)

func Inspect() {
	const meshName = "inspect"

	BeforeAll(func() {
		Expect(env.Global.Install(MTLSMeshUniversal(meshName))).To(Succeed())
		Expect(WaitForMesh(meshName, env.Zones())).To(Succeed())

		err := env.UniZone1.Install(TestServerUniversal("test-server", meshName,
			WithArgs([]string{"echo", "--instance", "echo"}),
		))
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Global.DeleteMesh(meshName)).To(Succeed())
	})

	Context("Dataplane", func() {
		It("should execute config dump from Global CP", func() {
			Eventually(func(g Gomega) {
				out, err := env.Global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", "kuma-4.test-server", "--config-dump", "--mesh", meshName)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).Should(ContainSubstring(`"dataplane.proxyType": "dataplane"`))
			}, "30s", "1s").Should(Succeed())
		})

		It("should execute config dump from Zone CP", func() {
			Eventually(func(g Gomega) {
				out, err := env.UniZone1.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", "test-server", "--config-dump", "--mesh", meshName)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).Should(ContainSubstring(`"dataplane.proxyType": "dataplane"`))
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("ZoneIngress", func() {
		It("should execute config dump from Global CP", func() {
			Eventually(func(g Gomega) {
				out, err := env.Global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zoneingress", "kuma-4.ingress", "--config-dump")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).Should(ContainSubstring(`"dataplane.proxyType": "ingress"`))
			}, "30s", "1s").Should(Succeed())
		})

		It("should execute config dump from Zone CP", func() {
			Eventually(func(g Gomega) {
				out, err := env.UniZone1.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zoneingress", "ingress", "--config-dump")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).Should(ContainSubstring(`"dataplane.proxyType": "ingress"`))
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("ZoneEgress", func() {
		It("should execute config dump from Global CP", func() {
			Eventually(func(g Gomega) {
				out, err := env.Global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zoneegress", "kuma-4.egress", "--config-dump")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).Should(ContainSubstring(`"dataplane.proxyType": "egress"`))
			}, "30s", "1s").Should(Succeed())
		})

		It("should execute config dump from Zone CP", func() {
			Eventually(func(g Gomega) {
				out, err := env.UniZone1.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zoneegress", "egress", "--config-dump")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).Should(ContainSubstring(`"dataplane.proxyType": "egress"`))
			}, "30s", "1s").Should(Succeed())
		})
	})
}
