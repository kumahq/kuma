package defaults

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func Defaults() {
	It("should create a Zone resource on every zone", func() {
		for _, zone := range multizone.Zones() {
			Eventually(func(g Gomega) {
				// when
				out, err := zone.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")

				// then
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).To(ContainSubstring(zone.ZoneName()))
				g.Expect(out).To(ContainSubstring("Online"))
			}, "30s", "1s").Should(Succeed())
		}
	})

	It("should create default HostnameGenerators every zone", func() {
		for _, zone := range multizone.Zones() {
			Eventually(func(g Gomega) {
				// when
				out, err := zone.GetKumactlOptions().RunKumactlAndGetOutput("get", "hostnamegenerators")

				// then
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).To(ContainSubstring("synced-"))
			}, "30s", "1s").Should(Succeed())
		}
	})

	It("should sync default zone HostnameGenerators to global for visibility", func() {
		Eventually(func(g Gomega) {
			// when
			out, err := multizone.Global.GetKumactlOptions().RunKumactlAndGetOutput("get", "hostnamegenerators")

			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(ContainSubstring("local-"))
		}, "30s", "1s").Should(Succeed())
	})
}
