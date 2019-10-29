package bootstrap

import (
	"github.com/Kong/kuma/pkg/config/api-server/catalogue"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auto configuration", func() {

	It("should auto configure catalogue with public settings for dataplane token server", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		cfg.General.AdvertisedHostname = "kuma.internal"
		cfg.DataplaneTokenServer.Local.Port = 1111
		cfg.DataplaneTokenServer.Public.Interface = "192.168.0.1"
		cfg.DataplaneTokenServer.Public.Port = 2222
		cfg.BootstrapServer.Port = 3333

		// when
		err := autoconfigure(&cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		expected := catalogue.CatalogueConfig{
			Bootstrap: catalogue.BootstrapApiConfig{
				Url: "http://kuma.internal:3333",
			},
			DataplaneToken: catalogue.DataplaneTokenApiConfig{
				LocalUrl:  "http://localhost:1111",
				PublicUrl: "https://kuma.internal:2222",
			},
		}
		Expect(*cfg.ApiServer.Catalogue).To(Equal(expected))
	})

	It("should auto configure catalogue without public settings for dataplane token server", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		cfg.General.AdvertisedHostname = "kuma.internal"
		cfg.DataplaneTokenServer.Local.Port = 1111
		cfg.BootstrapServer.Port = 3333

		// when
		err := autoconfigure(&cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		expected := catalogue.CatalogueConfig{
			Bootstrap: catalogue.BootstrapApiConfig{
				Url: "http://kuma.internal:3333",
			},
			DataplaneToken: catalogue.DataplaneTokenApiConfig{
				LocalUrl:  "http://localhost:1111",
				PublicUrl: "",
			},
		}
		Expect(*cfg.ApiServer.Catalogue).To(Equal(expected))
	})
})
