package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

var _ = Describe("DNSConfigurer", func() {
	type testCase struct {
		vips     map[string][]string
		expected string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewInboundListenerBuilder(envoy.APIV3, "192.168.0.1", 1234, xds.SocketAddressProtocolUDP).
				WithOverwriteName(names.GetDNSListenerName()).
				Configure(DNS(given.vips)).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic TCP listener", testCase{
			vips: map[string][]string{
				"something.mesh": {"240.0.0.0"},
				"something.com":  {"240.0.0.0"},
				"backend.mesh":   {"240.0.0.1", "::2"},
			},
			expected: `
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 1234
                protocol: UDP
            enableReusePort: true
            listenerFilters:
            - name: envoy.filters.udp.dns_filter
              typedConfig:
                '@type': type.googleapis.com/envoy.extensions.filters.udp.dns_filter.v3.DnsFilterConfig
                serverConfig:
                  inlineDnsTable:
                    virtualDomains:
                    - answerTtl: 30s
                      endpoint:
                        addressList:
                          address:
                          - 240.0.0.1
                          - ::2
                      name: backend.mesh
                    - answerTtl: 30s
                      endpoint:
                        addressList:
                          address:
                          - 240.0.0.0
                      name: something.com
                    - answerTtl: 30s
                      endpoint:
                        addressList:
                          address:
                          - 240.0.0.0
                      name: something.mesh
                statPrefix: kuma_dns
            name: kuma:dns
            trafficDirection: INBOUND
`,
		}),
	)
})
