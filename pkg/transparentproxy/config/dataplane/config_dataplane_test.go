package dataplane_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds/types"
	tproxy_config "github.com/kumahq/kuma/v2/pkg/transparentproxy/config"
	tproxy_dp "github.com/kumahq/kuma/v2/pkg/transparentproxy/config/dataplane"
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

	Describe("HasVNet", func() {
		It("should return false for nil config", func() {
			var cfg *tproxy_dp.DataplaneConfig
			Expect(cfg.HasVNet()).To(BeFalse())
		})

		It("should return false when no VNet networks are configured", func() {
			cfg := &tproxy_dp.DataplaneConfig{}
			Expect(cfg.HasVNet()).To(BeFalse())
		})

		It("should return true when VNet networks are configured", func() {
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

	Describe("VNet deserialization", func() {
		It("should deserialize VNet from full config YAML", func() {
			// This simulates what happens when kuma-dp loads the transparent proxy
			// config file written by `kumactl install transparent-proxy --store-config`.
			// The full config YAML contains redirect.vnet.networks, and the
			// non-strict YAML decoder populates matching fields in DataplaneConfig.
			cfgYAML := `
ipFamilyMode: dualstack
redirect:
  inbound:
    enabled: true
    port: 15006
  outbound:
    enabled: true
    port: 15001
  dns:
    enabled: true
    port: 15053
  vnet:
    networks:
    - "docker0:172.17.0.0/16"
`
			cfg := tproxy_dp.DataplaneConfig{}
			err := yaml.Unmarshal([]byte(cfgYAML), &cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.HasVNet()).To(BeTrue())
			Expect(cfg.Redirect.VNet.Networks).To(Equal([]string{"docker0:172.17.0.0/16"}))
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
				dnsPort: 12345,
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
