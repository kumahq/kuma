package bootstrap

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/config/multizone"
	core_manager "github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	kds_context "github.com/kumahq/kuma/v3/pkg/kds/context"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
)

var _ = Describe("KDS authentication", func() {
	configure := func(cfg kuma_cp.Config) (*kds_context.Context, error) {
		kdsContext := &kds_context.Context{}
		return kdsContext, configureKDSAuth(kdsContext, core_manager.NewResourceManager(memory.NewStore()), cfg)
	}

	zoneConfig := func(globalAddress string) kuma_cp.Config {
		cfg := kuma_cp.DefaultConfig()
		cfg.Mode = config_core.Zone
		cfg.Multizone.Zone.GlobalAddress = globalAddress
		cfg.Multizone.Zone.KDS.Auth.TokenInline = "token"
		return cfg
	}

	It("should be disabled by default", func() {
		for _, mode := range []config_core.CpMode{config_core.Global, config_core.Zone} {
			cfg := kuma_cp.DefaultConfig()
			cfg.Mode = mode
			cfg.Multizone.Zone.GlobalAddress = "grpcs://global:5685"

			kdsContext, err := configure(cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(kdsContext.GlobalZoneAuthenticator).To(BeNil())
			Expect(kdsContext.ZoneCredentials).To(BeNil())
		}
	})

	It("should authenticate zones on Global CP with zoneToken type", func() {
		cfg := kuma_cp.DefaultConfig()
		cfg.Mode = config_core.Global
		cfg.Multizone.Global.KDS.Auth.Type = multizone.KDSAuthZoneToken

		kdsContext, err := configure(cfg)

		Expect(err).ToNot(HaveOccurred())
		Expect(kdsContext.GlobalZoneAuthenticator).ToNot(BeNil())
		Expect(kdsContext.ZoneCredentials).To(BeNil())
	})

	It("should not authenticate zones on Zone CP", func() {
		cfg := kuma_cp.DefaultConfig()
		cfg.Mode = config_core.Zone
		cfg.Multizone.Global.KDS.Auth.Type = multizone.KDSAuthZoneToken

		kdsContext, err := configure(cfg)

		Expect(err).ToNot(HaveOccurred())
		Expect(kdsContext.GlobalZoneAuthenticator).To(BeNil())
	})

	It("should send the token from Zone CP over TLS", func() {
		kdsContext, err := configure(zoneConfig("grpcs://global:5685"))

		Expect(err).ToNot(HaveOccurred())
		Expect(kdsContext.ZoneCredentials).ToNot(BeNil())
		Expect(kdsContext.ZoneCredentials.RequireTransportSecurity()).To(BeTrue())
	})

	It("should refuse to send the token over plaintext", func() {
		_, err := configure(zoneConfig("grpc://global:5685"))

		Expect(err).To(MatchError(ContainSubstring("Use grpcs scheme")))
	})

	It("should ignore the token on a non federated Zone CP", func() {
		kdsContext, err := configure(zoneConfig(""))

		Expect(err).ToNot(HaveOccurred())
		Expect(kdsContext.ZoneCredentials).To(BeNil())
	})

	It("should fail when the token file cannot be read", func() {
		cfg := zoneConfig("grpcs://global:5685")
		cfg.Multizone.Zone.KDS.Auth.TokenPath = "/not-existing/token"

		_, err := configure(cfg)

		Expect(err).To(MatchError(ContainSubstring("could not read zone token")))
	})
})
