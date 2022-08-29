package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

var _ = Describe("TagsHeaderConfigurer", func() {

	type testCase struct {
		tags     mesh_proto.MultiValueTagSet
		expected string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			routeConfiguration, err := routes.NewRouteConfigurationBuilder(envoy.APIV3).
				Configure(routes.TagsHeader(given.tags)).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(routeConfiguration)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic tags", testCase{
			tags: map[string]map[string]bool{
				"tag2": {"value21": true},
				"tag3": {"value31": true, "value32": true, "value33": true},
				"tag1": {"value11": true, "value12": true},
			},
			expected: `
            requestHeadersToAdd:
            - header:
                key: x-kuma-tags
                value: '&tag1=value11,value12&&tag2=value21&&tag3=value31,value32,value33&'`,
		}),
		Entry("empty tags", testCase{
			tags:     map[string]map[string]bool{},
			expected: `{}`,
		}),
	)
})
