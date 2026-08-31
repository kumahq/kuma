package xds_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/plugin/xds"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var _ = Describe("Match order", func() {
	type validTestCase struct {
		conf          api.Conf
		orderedGolden string
	}
	DescribeTable("should generate proper order",
		func(given validTestCase) {
			// when
			orderedFilterChainMatches := plugin_xds.GetOrderedMatchers(given.conf)

			yaml, err := yaml.Marshal(orderedFilterChainMatches)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(yaml).To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s", given.orderedGolden)))
		},
		Entry("many different protocols", validTestCase{
			conf: api.Conf{
				AppendMatch: &[]api.Match{
					{
						Type:     api.MatchType("Domain"),
						Value:    "api.example.com",
						Port:     pointer.To[uint32](443),
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "api.example.com",
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "example.com",
						Port:     pointer.To[uint32](443),
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "*.example.com",
						Port:     pointer.To[uint32](443),
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "example.com",
						Port:     pointer.To[uint32](8080),
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "other.com",
						Port:     pointer.To[uint32](8080),
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "anotherhttp.com",
						Port:     pointer.To[uint32](8080),
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "http2.com",
						Port:     pointer.To[uint32](9000),
						Protocol: api.ProtocolType("http2"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "grpc.com",
						Port:     pointer.To[uint32](9001),
						Protocol: api.ProtocolType("grpc"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "*.example.com",
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "*.example.com",
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("IP"),
						Value:    "10.42.0.8",
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("IP"),
						Value:    "192.168.19.1",
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("IP"),
						Value:    "192.168.0.1",
						Port:     pointer.To[uint32](9091),
						Protocol: api.ProtocolType("tcp"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "otherexample.com",
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("CIDR"),
						Value:    "192.168.0.1/24",
						Protocol: api.ProtocolType("tcp"),
					},
					{
						Type:     api.MatchType("CIDR"),
						Value:    "192.168.1.1/30",
						Protocol: api.ProtocolType("tcp"),
					},
					{
						Type:     api.MatchType("CIDR"),
						Value:    "192.168.2.1/30",
						Protocol: api.ProtocolType("tcp"),
					},
					{
						Type:     api.MatchType("CIDR"),
						Value:    "192.168.0.1/30",
						Protocol: api.ProtocolType("tcp"),
					},
					{
						Type:     api.MatchType("CIDR"),
						Value:    "240.0.0.0/4",
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("CIDR"),
						Value:    "172.18.0.0/16",
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("IP"),
						Value:    "b6e5:a45e:70ae:e77f:d24e:5023:375d:20a6",
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("IP"),
						Value:    "9942:9abf:d0e0:f2da:2290:333b:e590:f497",
						Port:     pointer.To[uint32](9091),
						Protocol: api.ProtocolType("tcp"),
					},
					{
						Type:     api.MatchType("CIDR"),
						Value:    "b0ce:f616:4e74:28f7:427c:b969:8016:6344/64",
						Protocol: api.ProtocolType("tcp"),
					},
					{
						Type:     api.MatchType("CIDR"),
						Value:    "b0ce:f616:4e74:28f7:427c:b969:8016:6344/96",
						Protocol: api.ProtocolType("tcp"),
					},
				},
			},
			orderedGolden: "ordered.golden.yaml",
		}),
		Entry("different protocols on the same port but only one L7", validTestCase{
			conf: api.Conf{
				AppendMatch: &[]api.Match{
					{
						Type:     api.MatchType("Domain"),
						Value:    "api.example.com",
						Port:     pointer.To[uint32](443),
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "api.example.com",
						Port:     pointer.To[uint32](443),
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("IP"),
						Value:    "127.0.0.1",
						Port:     pointer.To[uint32](443),
						Protocol: api.ProtocolType("tcp"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "api.example.com",
						Port:     pointer.To[uint32](9090),
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "api.example.com",
						Port:     pointer.To[uint32](9090),
						Protocol: api.ProtocolType("http2"),
					},
					{
						Type:     api.MatchType("IP"),
						Value:    "127.0.0.1",
						Port:     pointer.To[uint32](9090),
						Protocol: api.ProtocolType("tcp"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "api.example.com",
						Port:     pointer.To[uint32](9091),
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "api.example.com",
						Port:     pointer.To[uint32](9091),
						Protocol: api.ProtocolType("grpc"),
					},
					{
						Type:     api.MatchType("IP"),
						Value:    "127.0.0.1",
						Port:     pointer.To[uint32](9091),
						Protocol: api.ProtocolType("tcp"),
					},
					{
						Type:     api.MatchType("IP"),
						Value:    "127.0.0.1",
						Protocol: api.ProtocolType("tcp"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "httpbin.com",
						Port:     pointer.To[uint32](80),
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("IP"),
						Value:    "10.22.22.1",
						Protocol: api.ProtocolType("http"),
					},
				},
			},
			orderedGolden: "ordered-diff-protocols.golden.yaml",
		}),
	)
	type conflictingTestCase struct {
		conf          api.Conf
		orderedGolden string
		warnings      []string
	}
	DescribeTable("should ignore conflicting matches instead of failing",
		func(given conflictingTestCase) {
			// when
			orderedFilterChainMatches := plugin_xds.GetOrderedMatchers(given.conf)

			yaml, err := yaml.Marshal(orderedFilterChainMatches)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(yaml).To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s", given.orderedGolden)))
			Expect(given.conf.Warnings()).To(Equal(given.warnings))
		},
		Entry("many different protocols", conflictingTestCase{
			conf: api.Conf{
				AppendMatch: &[]api.Match{
					{
						Type:     api.MatchType("Domain"),
						Value:    "example.com",
						Port:     pointer.To[uint32](8080),
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "another.com",
						Port:     pointer.To[uint32](8080),
						Protocol: api.ProtocolType("http2"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "other.com",
						Port:     pointer.To[uint32](8080),
						Protocol: api.ProtocolType("tcp"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "anotherhttp.com",
						Port:     pointer.To[uint32](9001),
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "http2.com",
						Port:     pointer.To[uint32](9001),
						Protocol: api.ProtocolType("http2"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "grpc.com",
						Port:     pointer.To[uint32](9001),
						Protocol: api.ProtocolType("grpc"),
					},
				},
			},
			orderedGolden: "conflicting-protocols-on-the-same-port.golden.yaml",
			warnings: []string{
				`ignoring match "another.com", protocols http and http2 produce the same filter chain for domains on port 8080`,
				`ignoring match "http2.com", protocols http and http2 produce the same filter chain for domains on port 9001`,
				`ignoring match "grpc.com", protocols http and grpc produce the same filter chain for domains on port 9001`,
			},
		}),
		Entry("the same domain on all ports and on a port with a different L7 protocol", conflictingTestCase{
			conf: api.Conf{
				AppendMatch: &[]api.Match{
					{
						Type:     api.MatchType("Domain"),
						Value:    "datadog.datadog.svc.cluster.local",
						Port:     pointer.To[uint32](4317),
						Protocol: api.ProtocolType("grpc"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "datadog.datadog.svc.cluster.local",
						Protocol: api.ProtocolType("http"),
					},
				},
			},
			orderedGolden: "conflicting-protocols-on-all-ports.golden.yaml",
			warnings: []string{
				"protocols grpc and http produce the same filter chain for domains on port 4317, matches with protocol http and no port are not applied there",
			},
		}),
		Entry("an IP and a CIDR resolving to the same address range with the same protocol", conflictingTestCase{
			conf: api.Conf{
				AppendMatch: &[]api.Match{
					{
						Type:     api.MatchType("IP"),
						Value:    "10.0.0.1",
						Port:     pointer.To[uint32](80),
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("CIDR"),
						Value:    "10.0.0.1/32",
						Port:     pointer.To[uint32](80),
						Protocol: api.ProtocolType("http"),
					},
				},
			},
			orderedGolden: "duplicate-ip-and-cidr.golden.yaml",
			warnings: []string{
				`ignoring match "10.0.0.1/32", matches "10.0.0.1" and "10.0.0.1/32" produce the same filter chain for 10.0.0.1/32 on port 80`,
			},
		}),
		Entry("tcp and mysql on the same address and port", conflictingTestCase{
			conf: api.Conf{
				AppendMatch: &[]api.Match{
					{
						Type:     api.MatchType("IP"),
						Value:    "10.1.1.1",
						Port:     pointer.To[uint32](3306),
						Protocol: api.ProtocolType("tcp"),
					},
					{
						Type:     api.MatchType("IP"),
						Value:    "10.1.1.1",
						Port:     pointer.To[uint32](3306),
						Protocol: api.ProtocolType("mysql"),
					},
				},
			},
			orderedGolden: "duplicate-tcp-and-mysql.golden.yaml",
			warnings: []string{
				`ignoring match "10.1.1.1", protocols tcp and mysql produce the same filter chain for 10.1.1.1/32 on port 3306`,
			},
		}),
		Entry("CIDRs with host bits resolving to the same range with different L7 protocols", conflictingTestCase{
			conf: api.Conf{
				AppendMatch: &[]api.Match{
					{
						Type:     api.MatchType("CIDR"),
						Value:    "10.0.0.1/24",
						Port:     pointer.To[uint32](8080),
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("CIDR"),
						Value:    "10.0.0.0/24",
						Port:     pointer.To[uint32](8080),
						Protocol: api.ProtocolType("grpc"),
					},
				},
			},
			orderedGolden: "duplicate-cidr-host-bits.golden.yaml",
			warnings: []string{
				`ignoring match "10.0.0.0/24", protocols http and grpc produce the same filter chain for 10.0.0.0/24 on port 8080`,
			},
		}),
		Entry("a match without a port duplicating an explicit chain of the same protocol", conflictingTestCase{
			conf: api.Conf{
				AppendMatch: &[]api.Match{
					{
						Type:     api.MatchType("IP"),
						Value:    "192.168.0.1",
						Protocol: api.ProtocolType("tcp"),
					},
					{
						Type:     api.MatchType("CIDR"),
						Value:    "192.168.0.1/32",
						Port:     pointer.To[uint32](9090),
						Protocol: api.ProtocolType("tcp"),
					},
				},
			},
			orderedGolden: "duplicate-all-ports-and-explicit-port.golden.yaml",
			warnings: []string{
				`matches "192.168.0.1/32" and "192.168.0.1" produce the same filter chain for 192.168.0.1/32 on port 9090, matches with protocol tcp and no port are not applied there`,
			},
		}),
		Entry("IPv6 spelled differently resolving to the same address", conflictingTestCase{
			conf: api.Conf{
				AppendMatch: &[]api.Match{
					{
						Type:     api.MatchType("IP"),
						Value:    "0:0:0:0:0:0:0:1",
						Port:     pointer.To[uint32](443),
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("CIDR"),
						Value:    "::1/128",
						Port:     pointer.To[uint32](443),
						Protocol: api.ProtocolType("tls"),
					},
				},
			},
			orderedGolden: "duplicate-ipv6-forms.golden.yaml",
			warnings: []string{
				`ignoring match "::1/128", matches "0:0:0:0:0:0:0:1" and "::1/128" produce the same filter chain for ::1/128 on port 443`,
			},
		}),
	)
})
