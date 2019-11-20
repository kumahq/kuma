package bootstrap

import (
	"github.com/Kong/kuma/pkg/config/api-server/catalog"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/config/core"
	gui_server "github.com/Kong/kuma/pkg/config/gui-server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auto configuration", func() {

	type testCase struct {
		cpConfig              func() kuma_cp.Config
		expectedCatalogConfig catalog.CatalogConfig
	}
	DescribeTable("should autoconfigure catalog",
		func(given testCase) {
			// given
			cfg := given.cpConfig()

			// when
			err := autoconfigure(&cfg)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(*cfg.ApiServer.Catalog).To(Equal(given.expectedCatalogConfig))
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
			expectedCatalogConfig: catalog.CatalogConfig{
				Bootstrap: catalog.BootstrapApiConfig{
					Url: "http://kuma.internal:3333",
				},
				DataplaneToken: catalog.DataplaneTokenApiConfig{
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
			expectedCatalogConfig: catalog.CatalogConfig{
				Bootstrap: catalog.BootstrapApiConfig{
					Url: "http://kuma.internal:3333",
				},
				DataplaneToken: catalog.DataplaneTokenApiConfig{
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
			expectedCatalogConfig: catalog.CatalogConfig{
				Bootstrap: catalog.BootstrapApiConfig{
					Url: "http://kuma.internal:3333",
				},
				DataplaneToken: catalog.DataplaneTokenApiConfig{
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
			expectedCatalogConfig: catalog.CatalogConfig{
				Bootstrap: catalog.BootstrapApiConfig{
					Url: "http://localhost:5682",
				},
				DataplaneToken: catalog.DataplaneTokenApiConfig{
					LocalUrl:  "",
					PublicUrl: "",
				},
			},
		}),
	)

	It("should autoconfigure gui config", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		cfg.Environment = core.KubernetesEnvironment
		cfg.General.AdvertisedHostname = "kuma.internal"
		cfg.ApiServer.Port = 1234

		// when
		err := autoconfigure(&cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(*cfg.GuiServer.GuiConfig).To(Equal(gui_server.GuiConfig{
			ApiUrl:      "http://kuma.internal:1234",
			Environment: "kubernetes",
		}))
	})

	It("should autoconfigure xds params", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		cfg.General.AdvertisedHostname = "kuma.internal"
		cfg.XdsServer.GrpcPort = 1234

		// when
		err := autoconfigure(&cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.BootstrapServer.Params.XdsHost).To(Equal("kuma.internal"))
		Expect(cfg.BootstrapServer.Params.XdsPort).To(Equal(uint32(1234)))
	})
})
