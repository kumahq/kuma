package cni

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewIntermediateConfig", func() {
	It("should set UID to default value if annotation is not specified", func() {
		a := map[string]string{}
		cfg, err := NewIntermediateConfig(a)
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.noRedirectUID).To(Equal(defaultNoRedirectUID))
	})

	It("should set ip family mode to dualstack by default", func() {
		a := map[string]string{}
		cfg, err := NewIntermediateConfig(a)
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.ipFamilyMode).To(Equal("dualstack"))
	})

	It("should override UID when annotation is specified", func() {
		a := map[string]string{
			"kuma.io/sidecar-uid": "1234",
		}
		cfg, err := NewIntermediateConfig(a)
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.noRedirectUID).To(Equal("1234"))
	})

	It("should set inbound redirect port for ipv4 and ipv6", func() {
		a := map[string]string{
			"kuma.io/transparent-proxying-inbound-port": "1234",
		}
		cfg, err := NewIntermediateConfig(a)
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.inboundPort).To(Equal("1234"))
	})

	It("should exclude application probe proxy ports", func() {
		a := map[string]string{
			"kuma.io/application-probe-proxy-port":  "19988",
			"traffic.kuma.io/exclude-inbound-ports": "3355",
		}
		cfg, err := NewIntermediateConfig(a)
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.excludeInboundPorts).To(Equal("3355,19988"))
	})
})
