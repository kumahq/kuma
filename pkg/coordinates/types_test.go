package coordinates_test

import (
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/coordinates"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FromConfig()", func() {

	var cfg kuma_cp.Config

	BeforeEach(func() {
		cfg = kuma_cp.DefaultConfig()
		cfg.Hostname = "kuma.internal"
		cfg.DataplaneTokenServer.Local.Port = 1111
		cfg.BootstrapServer.Port = 3333
	})

	It("should build coordinates from configuration with tls dataplane token server", func() {
		// given
		cfg.DataplaneTokenServer.Public.Interface = "192.168.0.1"
		cfg.DataplaneTokenServer.Public.Port = 2222

		// when
		coords := coordinates.FromConfig(cfg)

		// then
		expected := coordinates.Coordinates{
			Apis: coordinates.Apis{
				Bootstrap: coordinates.BootstrapApi{
					Url: "http://kuma.internal:3333",
				},
				DataplaneToken: coordinates.DataplaneTokenApi{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "https://kuma.internal:2222",
				},
			},
		}
		Expect(coords).To(Equal(expected))
	})

	It("should build coordinates from configuration without public dataplane token server", func() {
		// when
		coords := coordinates.FromConfig(cfg)

		// then
		expected := coordinates.Coordinates{
			Apis: coordinates.Apis{
				Bootstrap: coordinates.BootstrapApi{
					Url: "http://kuma.internal:3333",
				},
				DataplaneToken: coordinates.DataplaneTokenApi{
					LocalUrl:  "http://localhost:1111",
					PublicUrl: "",
				},
			},
		}
		Expect(coords).To(Equal(expected))
	})
})
