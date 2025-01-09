package xds_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/plugin/xds"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

var _ = Describe("Match order", func() {
	type validTestCase struct {
		conf          api.Conf
		orderedGolden string
	}
	DescribeTable("should generate proper order",
		func(given validTestCase) {
			// when
			orderedFilterChainMatches, _ := plugin_xds.GetOrderedMatchers(given.conf)

			yaml, err := yaml.Marshal(orderedFilterChainMatches)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(yaml).To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s", given.orderedGolden)))
		},
		Entry("many different protocols", validTestCase{
			conf: api.Conf{
				AppendMatch: []api.Match{
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
				AppendMatch: []api.Match{
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
	type invalidTestCase struct {
		conf      api.Conf
		errorMsgs []string
	}
	DescribeTable("should fail when many protocols L7 on the same port",
		func(given invalidTestCase) {
			// when
			_, err := plugin_xds.GetOrderedMatchers(given.conf)

			// then
			Expect(err).To(HaveOccurred())
			for _, errorMsg := range given.errorMsgs {
				Expect(err.Error()).To(ContainSubstring(errorMsg))
			}
			Expect(strings.Split(err.Error(), ";")).To(HaveLen(len(given.errorMsgs)))
		},
		Entry("many different protocols", invalidTestCase{
			conf: api.Conf{
				AppendMatch: []api.Match{
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
			errorMsgs: []string{"you cannot configure http, http2, grpc on the same port 8080", "you cannot configure http, http2, grpc on the same port 9001"},
		}),
	)
})
