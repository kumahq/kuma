package xds_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/plugin/xds"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

var _ = Describe("Match order", func() {
	type sidecarTestCase struct {
		conf          api.Conf
		orderedGolden string
	}
	DescribeTable("should generate proper order",
		func(given sidecarTestCase) {
			// when
			ordered, _ := plugin_xds.GetOrderedMatchers(given.conf)

			yaml, err := yaml.Marshal(ordered)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(yaml).To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s", given.orderedGolden)))
		},
		Entry("1", sidecarTestCase{
			conf: api.Conf{
				AppendMatch: []api.Match{
					{
						Type:     api.MatchType("Domain"),
						Value:    "api.example.com",
						Port:     pointer.To[int](443),
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
						Port:     pointer.To[int](443),
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "*.example.com",
						Port:     pointer.To[int](443),
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "example.com",
						Port:     pointer.To[int](8080),
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "other.com",
						Port:     pointer.To[int](8080),
						Protocol: api.ProtocolType("http"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "http2.com",
						Port:     pointer.To[int](8080),
						Protocol: api.ProtocolType("http2"),
					},
					{
						Type:     api.MatchType("Domain"),
						Value:    "*.example.com",
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("IP"),
						Value:    "192.168.19.1",
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("IP"),
						Value:    "192.168.0.1",
						Port:     pointer.To[int](9091),
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
						Type:     api.MatchType("IP"),
						Value:    "b6e5:a45e:70ae:e77f:d24e:5023:375d:20a6",
						Protocol: api.ProtocolType("tls"),
					},
					{
						Type:     api.MatchType("IP"),
						Value:    "9942:9abf:d0e0:f2da:2290:333b:e590:f497",
						Port:     pointer.To[int](9091),
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
	)
})
