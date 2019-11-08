package bootstrap

import (
	"github.com/Kong/kuma/pkg/config/api-server/catalogue"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auto configuration", func() {

	type testCase struct {
		cpConfig                func() kuma_cp.Config
		expectedCatalogueConfig catalogue.CatalogueConfig
	}
	DescribeTable("should autoconfigure catalogue",
		func(given testCase) {
			// given
			cfg := given.cpConfig()

			// when
			err := autoconfigure(&cfg)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(*cfg.ApiServer.Catalogue).To(Equal(given.expectedCatalogueConfig))
		},
		Entry("with public settings for dataplane token server", testCase{
			cpConfig: func() kuma_cp.Config {
				cfg := kuma_cp.DefaultConfig()
				cfg.General.AdvertisedHostname = "kuma.internal"
				cfg.DataplaneTokenServer.Local.Port = 1111
				cfg.DataplaneTokenServer.Public.Enabled = true
				cfg.DataplaneTokenServer.Public.Interface = "192.168.0.1"
				cfg.DataplaneTokenServer.Public.Port = 2222
				cfg.BootstrapServer.Port = 3333
				return cfg
			},
			expectedCatalogueConfig: catalogue.CatalogueConfig{
				Bootstrap: catalogue.BootstrapApiConfig{
					Url: "http://kuma.internal:3333",
				},
				DataplaneToken: catalogue.DataplaneTokenApiConfig{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "https://kuma.internal:2222",
				},
			},
		}),
		Entry("without public port explicitly defined", testCase{
			cpConfig: func() kuma_cp.Config {
				cfg := kuma_cp.DefaultConfig()
				cfg.General.AdvertisedHostname = "kuma.internal"
				cfg.DataplaneTokenServer.Local.Port = 1111
				cfg.DataplaneTokenServer.Public.Enabled = true
				cfg.DataplaneTokenServer.Public.Interface = "192.168.0.1"
				cfg.BootstrapServer.Port = 3333
				return cfg
			},
			expectedCatalogueConfig: catalogue.CatalogueConfig{
				Bootstrap: catalogue.BootstrapApiConfig{
					Url: "http://kuma.internal:3333",
				},
				DataplaneToken: catalogue.DataplaneTokenApiConfig{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "https://kuma.internal:1111", // port is autoconfigured from the local port
				},
			},
		}),
		Entry("without public settings for dataplane token server", testCase{
			cpConfig: func() kuma_cp.Config {
				cfg := kuma_cp.DefaultConfig()
				cfg.General.AdvertisedHostname = "kuma.internal"
				cfg.DataplaneTokenServer.Local.Port = 1111
				cfg.BootstrapServer.Port = 3333
				return cfg
			},
			expectedCatalogueConfig: catalogue.CatalogueConfig{
				Bootstrap: catalogue.BootstrapApiConfig{
					Url: "http://kuma.internal:3333",
				},
				DataplaneToken: catalogue.DataplaneTokenApiConfig{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "",
				},
			},
		}),
		Entry("without dataplane token server", testCase{
			cpConfig: func() kuma_cp.Config {
				cfg := kuma_cp.DefaultConfig()
				cfg.DataplaneTokenServer.Enabled = false
				return cfg
			},
			expectedCatalogueConfig: catalogue.CatalogueConfig{
				Bootstrap: catalogue.BootstrapApiConfig{
					Url: "http://localhost:5682",
				},
				DataplaneToken: catalogue.DataplaneTokenApiConfig{
					LocalUrl:  "",
					PublicUrl: "",
				},
			},
		}),
	)
})
