package dataplane_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	tproxy_config "github.com/kumahq/kuma/v3/pkg/transparentproxy/config"
	tproxy_dp "github.com/kumahq/kuma/v3/pkg/transparentproxy/config/dataplane"
)

type dummyMeta struct {
	transparentProxy *tproxy_dp.DataplaneConfig
	dnsPort          uint32
}

func (d *dummyMeta) GetTransparentProxy() *tproxy_dp.DataplaneConfig { return d.transparentProxy }

func (d *dummyMeta) GetDNSPort() uint32 { return d.dnsPort }

type dummyDP struct {
	address string
}

func (d *dummyDP) GetAddress() string { return d.address }

var _ = Describe("DataplaneConfig functions", func() {
	Describe("Enabled", func() {
		It("should return true if IPv4 and redirect is enabled", func() {
			// given
			dp := &dummyDP{address: "192.0.2.1"}
			meta := &dummyMeta{
				transparentProxy: &tproxy_dp.DataplaneConfig{
					IPFamilyMode: tproxy_config.IPFamilyModeDualStack,
					Redirect: tproxy_dp.DataplaneRedirect{
						Inbound:  tproxy_dp.DatalpaneTrafficFlow{Enabled: true},
						Outbound: tproxy_dp.DatalpaneTrafficFlow{Enabled: true},
					},
				},
			}

			// when
			cfg := tproxy_dp.GetDataplaneConfig(dp, meta)

			// then
			Expect(cfg.Enabled()).To(BeTrue())
		})

		It("should return false if address is not IPv4 and mode is IPv4", func() {
			// given
			dp := &dummyDP{address: "::1"}
			meta := &dummyMeta{
				transparentProxy: &tproxy_dp.DataplaneConfig{
					IPFamilyMode: tproxy_config.IPFamilyModeIPv4,
					Redirect: tproxy_dp.DataplaneRedirect{
						Inbound:  tproxy_dp.DatalpaneTrafficFlow{Enabled: true},
						Outbound: tproxy_dp.DatalpaneTrafficFlow{Enabled: true},
					},
				},
			}

			// when
			cfg := tproxy_dp.GetDataplaneConfig(dp, meta)

			// then
			Expect(cfg.Enabled()).To(BeFalse())
		})
	})

	Describe("EnabledIPv6", func() {
		It("should return true for mode not IPv4", func() {
			// given
			meta := &dummyMeta{
				transparentProxy: &tproxy_dp.DataplaneConfig{
					IPFamilyMode: tproxy_config.IPFamilyModeDualStack,
				},
			}

			// when
			cfg := tproxy_dp.GetDataplaneConfig(nil, meta)

			// then
			Expect(cfg.EnabledIPv6()).To(BeTrue())
		})
	})

	Describe("HasVNet", func() {
		It("should return false for nil config", func() {
			var cfg *tproxy_dp.DataplaneConfig
			Expect(cfg.HasVNet()).To(BeFalse())
		})

		It("should return false when no networks are configured", func() {
			cfg := &tproxy_dp.DataplaneConfig{}
			Expect(cfg.HasVNet()).To(BeFalse())
		})

		It("should return true when networks are configured", func() {
			cfg := &tproxy_dp.DataplaneConfig{
				Redirect: tproxy_dp.DataplaneRedirect{
					VNet: tproxy_dp.DataplaneVNet{
						Networks: []string{"docker0:172.17.0.0/16"},
					},
				},
			}
			Expect(cfg.HasVNet()).To(BeTrue())
		})
	})

	Describe("GetDataplaneConfig", func() {
		It("should return fallback if nil", func() {
			cfg := tproxy_dp.GetDataplaneConfig(nil, nil)
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Enabled()).To(BeFalse())
		})

		It("should use meta and dp values", func() {
			// given
			dp := &dummyDP{address: "192.0.2.100"}

			meta := &dummyMeta{
				transparentProxy: &tproxy_dp.DataplaneConfig{
					IPFamilyMode: tproxy_config.IPFamilyModeDualStack,
					Redirect: tproxy_dp.DataplaneRedirect{
						Inbound:  tproxy_dp.DatalpaneTrafficFlow{Enabled: true},
						Outbound: tproxy_dp.DatalpaneTrafficFlow{Enabled: true},
					},
				},
				dnsPort: 12345,
			}

			// when
			cfg := tproxy_dp.GetDataplaneConfig(dp, meta)

			// then
			Expect(cfg.Redirect.DNS.Port.Uint32()).To(Equal(uint32(12345)))
			Expect(cfg.Enabled()).To(BeTrue())
		})

		It("should return a disabled config when metadata carries no transparent proxy", func() {
			// given
			dp := &dummyDP{address: "192.0.2.50"}
			meta := &dummyMeta{dnsPort: 12345}

			// when
			cfg := tproxy_dp.GetDataplaneConfig(dp, meta)

			// then
			Expect(cfg.Redirect.DNS.Port.Uint32()).To(Equal(uint32(12345)))
			Expect(cfg.Enabled()).To(BeFalse())
		})
	})
})
