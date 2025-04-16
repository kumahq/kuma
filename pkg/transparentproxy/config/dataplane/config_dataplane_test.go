package dataplane_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/pkg/core/xds/types"
	tproxy_config "github.com/kumahq/kuma/pkg/transparentproxy/config"
	tproxy_dp "github.com/kumahq/kuma/pkg/transparentproxy/config/dataplane"
)

type dummyMeta struct {
	tproxy_dp.DataplaneConfig
	features map[string]bool
	dnsPort  uint32
}

func (d *dummyMeta) GetTransparentProxy() *tproxy_dp.DataplaneConfig { return &d.DataplaneConfig }

func (d *dummyMeta) HasFeature(f string) bool { return d.features[f] }

func (d *dummyMeta) GetDNSPort() uint32 { return d.dnsPort }

type dummyDP struct {
	tproxy_dp.DataplaneConfig
	address string
}

func (d *dummyDP) GetTransparentProxy() *tproxy_dp.DataplaneConfig { return &d.DataplaneConfig }

func (d *dummyDP) GetAddress() string { return d.address }

var _ = Describe("DataplaneConfig functions", func() {
	Describe("Enabled", func() {
		It("should return true if IPv4 and redirect is enabled", func() {
			// given
			dp := &dummyDP{
				DataplaneConfig: tproxy_dp.DataplaneConfig{
					IPFamilyMode: tproxy_config.IPFamilyModeDualStack,
					Redirect: tproxy_dp.DataplaneRedirect{
						Inbound:  tproxy_dp.DatalpaneTrafficFlow{Enabled: true},
						Outbound: tproxy_dp.DatalpaneTrafficFlow{Enabled: true},
					},
				},
				address: "192.0.2.1",
			}

			// when
			cfg := tproxy_dp.GetDataplaneConfig(dp, nil)

			// then
			Expect(cfg.Enabled()).To(BeTrue())
		})

		It("should return false if address is not IPv4 and mode is IPv4", func() {
			// given
			dp := &dummyDP{
				DataplaneConfig: tproxy_dp.DataplaneConfig{
					IPFamilyMode: tproxy_config.IPFamilyModeIPv4,
					Redirect: tproxy_dp.DataplaneRedirect{
						Inbound:  tproxy_dp.DatalpaneTrafficFlow{Enabled: true},
						Outbound: tproxy_dp.DatalpaneTrafficFlow{Enabled: true},
					},
				},
				address: "::1",
			}

			// when
			cfg := tproxy_dp.GetDataplaneConfig(dp, nil)

			// then
			Expect(cfg.Enabled()).To(BeFalse())
		})
	})

	Describe("EnabledIPv6", func() {
		It("should return true for mode not IPv4", func() {
			// given
			dp := &dummyDP{
				DataplaneConfig: tproxy_dp.DataplaneConfig{
					IPFamilyMode: tproxy_config.IPFamilyModeDualStack,
				},
			}

			// when
			cfg := tproxy_dp.GetDataplaneConfig(dp, nil)

			// then
			Expect(cfg.EnabledIPv6()).To(BeTrue())
		})
	})

	Describe("GetDataplaneConfig", func() {
		It("should return fallback if nil", func() {
			cfg := tproxy_dp.GetDataplaneConfig(nil, nil)
			Expect(cfg).ToNot(BeNil())
		})

		It("should use meta and dp values", func() {
			// given
			dp := &dummyDP{address: "192.0.2.100"}

			meta := &dummyMeta{
				DataplaneConfig: tproxy_dp.DataplaneConfig{
					IPFamilyMode: tproxy_config.IPFamilyModeDualStack,
					Redirect: tproxy_dp.DataplaneRedirect{
						Inbound:  tproxy_dp.DatalpaneTrafficFlow{Enabled: true},
						Outbound: tproxy_dp.DatalpaneTrafficFlow{Enabled: true},
					},
				},
				features: map[string]bool{
					core_xds.FeatureTransparentProxyInDataplaneMetadata: true,
				},
				dnsPort:  12345,
			}

			// when
			cfg := tproxy_dp.GetDataplaneConfig(dp, meta)

			// then
			Expect(cfg.Redirect.DNS.Port.Uint32()).To(Equal(uint32(12345)))
			Expect(cfg.Enabled()).To(BeTrue())
		})

		It("should fallback to dataplane config", func() {
			// given
			dp := &dummyDP{
				DataplaneConfig: tproxy_dp.DataplaneConfig{
					IPFamilyMode: tproxy_config.IPFamilyModeDualStack,
					Redirect: tproxy_dp.DataplaneRedirect{
						Inbound:  tproxy_dp.DatalpaneTrafficFlow{Enabled: true},
						Outbound: tproxy_dp.DatalpaneTrafficFlow{Enabled: true},
					},
				},
				address: "192.0.2.50",
			}

			// when
			cfg := tproxy_dp.GetDataplaneConfig(dp, nil)

			// then
			Expect(cfg.Redirect.DNS.Port.Uint32()).To(Equal(uint32(0)))
			Expect(cfg.Enabled()).To(BeTrue())
		})
	})
})
