package api_server_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/v3/pkg/config/api-server"
	config_types "github.com/kumahq/kuma/v3/pkg/config/types"
)

var _ = Describe("ApiServerConfig", func() {
	It("accumulates validation errors in configuration order", func() {
		cfg := api_server.DefaultApiServerConfig()
		cfg.HTTP.Interface = ""
		cfg.HTTPS.Interface = ""
		cfg.GUI.RootUrl = "%"
		cfg.RootUrl = "%"
		cfg.BasePath = "%"
		cfg.Authn.Tokens.Validator.PublicKeys = []config_types.PublicKey{{}}
		cfg.ReadHeaderTimeout.Duration = -time.Second
		cfg.ReadTimeout.Duration = -time.Second
		cfg.WriteTimeout.Duration = -time.Second
		cfg.IdleTimeout.Duration = -time.Second

		Expect(cfg.Validate()).To(MatchError(
			".HTTP not valid: Interface cannot be empty; " +
				".HTTPS not valid: .Interface cannot be empty; " +
				".GUI not valid: RootUrl is not a valid url; " +
				"RootUrl is not a valid URL; " +
				"BaseGuiPath is not a valid url; " +
				".Authn is not valid: .Tokens is not valid: .Validator is not valid: .PublicKeys[0] is not valid: .KID is required; " +
				".ReadHeaderTimeout must be greater or equal 0s; " +
				".ReadTimeout must be greater or equal 0s; " +
				".WriteTimeout must be greater or equal 0s; " +
				".IdleTimeout must be greater or equal 0s",
		))
	})

	It("accepts the default configuration", func() {
		Expect(api_server.DefaultApiServerConfig().Validate()).To(Succeed())
	})
})
