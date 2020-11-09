package bootstrap

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/api-server/catalog"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
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
		Entry("with default settings", testCase{
			cpConfig: func() kuma_cp.Config {
				cfg := kuma_cp.DefaultConfig()
				cfg.General.AdvertisedHostname = "kuma.internal"
				return cfg
			},
			expectedCatalogConfig: catalog.CatalogConfig{
				ApiServer: catalog.ApiServerConfig{
					Url: "http://kuma.internal:5681",
				},
				Bootstrap: catalog.BootstrapApiConfig{
					Url: "https://kuma.internal:5678",
				},
				DataplaneToken: catalog.DataplaneTokenApiConfig{
					LocalUrl: "http://localhost:5681",
				},
				MonitoringAssignment: catalog.MonitoringAssignmentApiConfig{
					Url: "grpc://kuma.internal:5676",
				},
			},
		}),
	)

	It("should autoconfigure xds params", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		cfg.General.AdvertisedHostname = "kuma.internal"
		cfg.DpServer.Port = 1234

		// when
		err := autoconfigure(&cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.BootstrapServer.Params.XdsHost).To(Equal("kuma.internal"))
		Expect(cfg.BootstrapServer.Params.XdsPort).To(Equal(uint32(1234)))
	})
})
