package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

var _ = Describe("DNSConfigurer", func() {

	type testCase struct {
		vips         map[string][]string
		emptyDnsPort uint32
		envoyVersion string
		expected     string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder(envoy.APIV3).
				Configure(InboundListener(names.GetDNSListenerName(), "192.168.0.1", 1234, xds.SocketAddressProtocolUDP)).
				Configure(DNS(given.vips, given.emptyDnsPort, &mesh_proto.EnvoyVersion{Version: given.envoyVersion})).
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
			emptyDnsPort: 53002,
			envoyVersion: "1.20.0",
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
                clientConfig:
                  maxPendingLookups: "256"
                  typedDnsResolverConfig:
                      name: envoy.network.dns_resolver.cares
                      typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.network.dns_resolver.cares.v3.CaresDnsResolverConfig
                          resolvers:
                              - socketAddress:
                                  address: 127.0.0.1
                                  portValue: 53002
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
		Entry("basic TCP listener, Envoy 1.18.4", testCase{
			vips: map[string][]string{
				"something.mesh": {"240.0.0.0"},
				"something.com":  {"240.0.0.0"},
				"backend.mesh":   {"240.0.0.1", "::2"},
			},
			emptyDnsPort: 53002,
			envoyVersion: "1.18.4",
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
                '@type': type.googleapis.com/envoy.extensions.filters.udp.dns_filter.v3alpha.DnsFilterConfig
                clientConfig:
                  maxPendingLookups: "256"
                  upstreamResolvers:
                  - socketAddress:
                      address: 127.0.0.1
                      portValue: 53002
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
