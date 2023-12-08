package defaults

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func Defaults() {
	It("should create a Zone resource on every zone", func() {
		for _, zone := range multizone.Zones() {
			Eventually(func(g Gomega) {
				// when
				out, err := zone.GetKumactlOptions().RunKumactlAndGetOutput("get", "zone", zone.ZoneName(), "-ojson")

				// then
				g.Expect(err).ToNot(HaveOccurred())
				meta := &v1alpha1.ResourceMeta{}
				g.Expect(json.Unmarshal([]byte(out), meta)).To(Succeed())
				g.Expect(meta.Name).To(Equal(zone.ZoneName()))
			}, "30s", "1s").Should(Succeed())
		}
	})
}
