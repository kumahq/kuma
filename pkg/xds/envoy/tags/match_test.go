package tags_test

import (
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/tags"
)

func strictMatch(re *regexp.Regexp, s string) bool {
	idx := re.FindStringIndex(s)
	if idx == nil {
		return false
	}
	return idx[0] == 0 && idx[1] == len(s)
}

var _ = Describe("MatchingRegex", func() {
	type testCase struct {
		serviceTags mesh_proto.MultiValueTagSet
		selector    mesh_proto.SingleValueTagSet
		expected    bool
	}

	DescribeTable("should generate regex for matching service's tags",
		func(given testCase) {
			// when
			regexStr := tags.MatchingRegex(given.selector)
			re, err := regexp.Compile(regexStr)
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			matched := strictMatch(re, tags.Serialize(given.serviceTags))
			// then
			Expect(matched).To(Equal(given.expected))
		},
		Entry("match 2 one value tags", testCase{
			serviceTags: mesh_proto.MultiValueTagSet{
				"tag1": {"value1": true},
				"tag2": {"value2": true},
			},
			selector: mesh_proto.SingleValueTagSet{
				"tag1": "value1",
				"tag2": "value2",
			},
			expected: true,
		}),
		Entry("match 3 one-value tags", testCase{
			serviceTags: mesh_proto.MultiValueTagSet{
				"tag1": {"value1": true},
				"tag2": {"value2": true},
				"tag3": {"value3": true},
			},
			selector: mesh_proto.SingleValueTagSet{
				"tag2": "value2",
				"tag3": "value3",
			},
			expected: true,
		}),
		Entry("match without middle tag2", testCase{
			serviceTags: mesh_proto.MultiValueTagSet{
				"tag1": {"value1": true, "value2": true},
				"tag2": {"value2": true, "value3": true},
				"tag3": {"value3": true, "value4": true},
			},
			selector: mesh_proto.SingleValueTagSet{
				"tag1": "value1",
				"tag3": "value3",
			},
			expected: true,
		}),
		Entry("match the latter valuer", testCase{
			serviceTags: mesh_proto.MultiValueTagSet{
				"tag1": {"value1": true, "value2": true},
			},
			selector: mesh_proto.SingleValueTagSet{
				"tag1": "value2",
			},
			expected: true,
		}),
		Entry("shouldn't match", testCase{
			serviceTags: mesh_proto.MultiValueTagSet{
				"tag1": {"value1": true, "value2": true},
				"tag2": {"value2": true, "value3": true},
				"tag3": {"value3": true, "value4": true},
			},
			selector: mesh_proto.SingleValueTagSet{
				"tag1": "value1",
				"tag3": "value5",
			},
			expected: false,
		}),
		Entry("shouldn't match value's prefix", testCase{
			serviceTags: mesh_proto.MultiValueTagSet{
				"tag1": {"value1": true, "value2": true},
			},
			selector: mesh_proto.SingleValueTagSet{
				"tag1": "val",
			},
			expected: false,
		}),
		Entry("should match asterisk tag", testCase{
			serviceTags: mesh_proto.MultiValueTagSet{
				"tag1": {"value1": true, "value2": true},
			},
			selector: mesh_proto.SingleValueTagSet{
				"tag1": "*",
			},
			expected: true,
		}),
		Entry("shouldn't match asterisk tag", testCase{
			serviceTags: mesh_proto.MultiValueTagSet{
				"tag1": {"value1": true, "value2": true},
			},
			selector: mesh_proto.SingleValueTagSet{
				"tag2": "*",
			},
			expected: false,
		}),
		Entry("should escape dot in tag's value", testCase{
			serviceTags: mesh_proto.MultiValueTagSet{
				"version": {"0x1": true},
			},
			selector: mesh_proto.SingleValueTagSet{
				"version": "0.1",
			},
			expected: false,
		}),
		Entry("should match dot in tag's value", testCase{
			serviceTags: mesh_proto.MultiValueTagSet{
				"version": {"0.1": true},
			},
			selector: mesh_proto.SingleValueTagSet{
				"version": "0.1",
			},
			expected: true,
		}),
	)
})

var _ = Describe("FromLegacyTargetRef", func() {
	type testCase struct {
		targetRef common_api.TargetRef
		expected  tags.Tags
	}

	DescribeTable("should ignore subset tags and return only service or match-all tags",
		func(given testCase) {
			result, ok := tags.FromLegacyTargetRef(given.targetRef)
			Expect(ok).To(BeTrue())
			Expect(result).To(Equal(given.expected))
		},
		Entry("MeshService", testCase{
			targetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Name: pointer.To("backend"),
			},
			expected: tags.Tags{mesh_proto.ServiceTag: "backend"},
		}),
		Entry("MeshServiceSubset ignores tags", testCase{
			targetRef: common_api.TargetRef{
				Kind: common_api.MeshServiceSubset,
				Name: pointer.To("backend"),
				Tags: &map[string]string{"version": "v1", mesh_proto.ZoneTag: "east"},
			},
			expected: tags.Tags{mesh_proto.ServiceTag: "backend"},
		}),
		Entry("Mesh", testCase{
			targetRef: common_api.TargetRef{
				Kind: common_api.Mesh,
			},
			expected: tags.Tags{mesh_proto.ServiceTag: mesh_proto.MatchAllTag},
		}),
		Entry("MeshSubset ignores tags", testCase{
			targetRef: common_api.TargetRef{
				Kind: common_api.MeshSubset,
				Tags: &map[string]string{"version": "v1", mesh_proto.ZoneTag: "east"},
			},
			expected: tags.Tags{mesh_proto.ServiceTag: mesh_proto.MatchAllTag},
		}),
	)

	It("should return false for unsupported kind", func() {
		_, ok := tags.FromLegacyTargetRef(common_api.TargetRef{Kind: common_api.Dataplane})
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("RegexOR", func() {
	type testCase struct {
		servicesTags []mesh_proto.MultiValueTagSet
		selectors    []mesh_proto.SingleValueTagSet
		expected     []bool
	}

	DescribeTable("should generate regex based on several Selectors",
		func(given testCase) {
			var rss []string
			for _, s := range given.selectors {
				rss = append(rss, tags.MatchingRegex(s))
			}
			regexOR := tags.RegexOR(rss...)
			re, err := regexp.Compile(regexOR)
			Expect(err).ToNot(HaveOccurred())

			for i, service := range given.servicesTags {
				matched := strictMatch(re, tags.Serialize(service))
				Expect(matched).To(Equal(given.expected[i]))
			}
		},
		Entry("should match 2 services of 3", testCase{
			servicesTags: []mesh_proto.MultiValueTagSet{
				{
					"kuma.io/service": {"web": true, "web-api": true},
				},
				{
					"kuma.io/service": {"backend": true},
					"version":         {"3": true},
				},
				{
					"kuma.io/service": {"backend": true},
					"version":         {"2": true},
				},
			},
			selectors: []mesh_proto.SingleValueTagSet{
				{
					"kuma.io/service": "web",
				},
				{
					"kuma.io/service": "backend",
					"version":         "3",
				},
			},
			expected: []bool{true, true, false},
		}),
	)
})
