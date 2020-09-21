package bootstrap

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/api-server/catalog"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/core"
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
		Entry("with public settings for admin server", testCase{
			cpConfig: func() kuma_cp.Config {
				cfg := kuma_cp.DefaultConfig()
				cfg.General.AdvertisedHostname = "kuma.internal"
				cfg.AdminServer.Local.Port = 1111
				cfg.AdminServer.Public.Enabled = true
				cfg.AdminServer.Public.Interface = "192.168.0.1"
				cfg.AdminServer.Public.Port = 2222
				cfg.BootstrapServer.Port = 3333
				cfg.ApiServer.Port = 1234
				return cfg
			},
			expectedCatalogConfig: catalog.CatalogConfig{
				ApiServer: catalog.ApiServerConfig{
					Url: "http://kuma.internal:1234",
				},
				Bootstrap: catalog.BootstrapApiConfig{
					Url: "https://kuma.internal:3333",
				},
				DataplaneToken: catalog.DataplaneTokenApiConfig{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "https://kuma.internal:2222",
				},
				Admin: catalog.AdminApiConfig{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "https://kuma.internal:2222",
				},
				MonitoringAssignment: catalog.MonitoringAssignmentApiConfig{
					Url: "grpc://kuma.internal:5676",
				},
			},
		}),
		Entry("without public port explicitly defined", testCase{
			cpConfig: func() kuma_cp.Config {
				cfg := kuma_cp.DefaultConfig()
				cfg.General.AdvertisedHostname = "kuma.internal"
				cfg.AdminServer.Local.Port = 1111
				cfg.AdminServer.Public.Enabled = true
				cfg.AdminServer.Public.Interface = "192.168.0.1"
				cfg.BootstrapServer.Port = 3333
				return cfg
			},
			expectedCatalogConfig: catalog.CatalogConfig{
				ApiServer: catalog.ApiServerConfig{
					Url: "http://kuma.internal:5681",
				},
				Bootstrap: catalog.BootstrapApiConfig{
					Url: "https://kuma.internal:3333",
				},
				DataplaneToken: catalog.DataplaneTokenApiConfig{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "https://kuma.internal:1111", // port is autoconfigured from the local port
				},
				Admin: catalog.AdminApiConfig{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "https://kuma.internal:1111", // port is autoconfigured from the local port
				},
				MonitoringAssignment: catalog.MonitoringAssignmentApiConfig{
					Url: "grpc://kuma.internal:5676",
				},
			},
		}),
		Entry("without public settings for admin server", testCase{
			cpConfig: func() kuma_cp.Config {
				cfg := kuma_cp.DefaultConfig()
				cfg.General.AdvertisedHostname = "kuma.internal"
				cfg.AdminServer.Local.Port = 1111
				cfg.BootstrapServer.Port = 3333
				return cfg
			},
			expectedCatalogConfig: catalog.CatalogConfig{
				ApiServer: catalog.ApiServerConfig{
					Url: "http://kuma.internal:5681",
				},
				Bootstrap: catalog.BootstrapApiConfig{
					Url: "https://kuma.internal:3333",
				},
				DataplaneToken: catalog.DataplaneTokenApiConfig{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "",
				},
				Admin: catalog.AdminApiConfig{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "",
				},
				MonitoringAssignment: catalog.MonitoringAssignmentApiConfig{
					Url: "grpc://kuma.internal:5676",
				},
			},
		}),
		Entry("without dataplane token server", testCase{
			cpConfig: func() kuma_cp.Config {
				cfg := kuma_cp.DefaultConfig()
				cfg.AdminServer.Apis.DataplaneToken.Enabled = false
				return cfg
			},
			expectedCatalogConfig: catalog.CatalogConfig{
				ApiServer: catalog.ApiServerConfig{
					Url: "http://localhost:5681",
				},
				Bootstrap: catalog.BootstrapApiConfig{
					Url: "https://localhost:5682",
				},
				DataplaneToken: catalog.DataplaneTokenApiConfig{
					LocalUrl:  "",
					PublicUrl: "",
				},
				Admin: catalog.AdminApiConfig{
					LocalUrl:  "http://localhost:5679",
					PublicUrl: "",
				},
				MonitoringAssignment: catalog.MonitoringAssignmentApiConfig{
					Url: "grpc://localhost:5676",
				},
			},
		}),
		Entry("with public settings for bootstrap and mads server", testCase{
			cpConfig: func() kuma_cp.Config {
				cfg := kuma_cp.DefaultConfig()
				cfg.General.AdvertisedHostname = "kuma.internal"
				cfg.AdminServer.Local.Port = 1111
				cfg.AdminServer.Public.Enabled = true
				cfg.AdminServer.Public.Interface = "192.168.0.1"
				cfg.AdminServer.Public.Port = 2222
				cfg.BootstrapServer.Port = 3333
				cfg.ApiServer.Catalog.Bootstrap.Url = "https://bootstrap.kuma.com:1234"
				cfg.ApiServer.Catalog.MonitoringAssignment.Url = "grpcs://mads.kuma.com:1234"
				return cfg
			},
			expectedCatalogConfig: catalog.CatalogConfig{
				ApiServer: catalog.ApiServerConfig{
					Url: "http://kuma.internal:5681",
				},
				Bootstrap: catalog.BootstrapApiConfig{
					Url: "https://bootstrap.kuma.com:1234",
				},
				DataplaneToken: catalog.DataplaneTokenApiConfig{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "https://kuma.internal:2222",
				},
				Admin: catalog.AdminApiConfig{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "https://kuma.internal:2222",
				},
				MonitoringAssignment: catalog.MonitoringAssignmentApiConfig{
					Url: "grpcs://mads.kuma.com:1234",
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
		Expect(cfg.GuiServer.ApiServerUrl).To(Equal(""))
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

	It("should autoconfigure MonitoringAssignment server", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		cfg.General.AdvertisedHostname = "kuma.internal"
		cfg.MonitoringAssignmentServer.GrpcPort = 8765

		// when
		err := autoconfigure(&cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.ApiServer.Catalog.MonitoringAssignment.Url).To(Equal("grpc://kuma.internal:8765"))
	})
})
