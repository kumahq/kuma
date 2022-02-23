package envoy_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ = Describe("Ingress Dataplane", func() {

	type testCase struct {
		tags         envoy.TagKeysSlice
		transformers []envoy.TagKeyTransformer
		expected     envoy.TagKeysSlice
	}
	DescribeTable("should transform TagKeySlices",
		func(given testCase) {
			res := given.tags.Transform(given.transformers...)
			Expect(given.expected).To(Equal(res))
		},
		Entry("empty", testCase{
			tags:         envoy.TagKeysSlice{},
			transformers: []envoy.TagKeyTransformer{},
			expected:     envoy.TagKeysSlice{},
		}),
		Entry("no transformer", testCase{
			tags:         envoy.TagKeysSlice{{"key"}},
			transformers: []envoy.TagKeyTransformer{},
			expected:     envoy.TagKeysSlice{envoy.TagKeys{"key"}},
		}),
		Entry("should sort tagKeys with no transformer", testCase{
			tags:         envoy.TagKeysSlice{{"key", "key2"}},
			transformers: []envoy.TagKeyTransformer{},
			expected:     envoy.TagKeysSlice{envoy.TagKeys{"key", "key2"}},
		}),
		Entry("order of tags slice doesn't matter", testCase{
			tags:         envoy.TagKeysSlice{{"key"}, {"key2"}},
			transformers: []envoy.TagKeyTransformer{},
			expected:     envoy.TagKeysSlice{envoy.TagKeys{"key2"}, envoy.TagKeys{"key"}}.Transform(),
		}),
		Entry("should dedupe duplicated set", testCase{
			tags:         envoy.TagKeysSlice{{"key", "key2"}, {"key"}, {"key"}},
			transformers: []envoy.TagKeyTransformer{},
			expected:     envoy.TagKeysSlice{envoy.TagKeys{"key"}, envoy.TagKeys{"key", "key2"}},
		}),
		Entry("should add key to all tags", testCase{
			tags:         envoy.TagKeysSlice{{"key2"}, {"key"}},
			transformers: []envoy.TagKeyTransformer{envoy.With("added1", "added2")},
			expected:     envoy.TagKeysSlice{envoy.TagKeys{"added1", "added2", "key"}, envoy.TagKeys{"added1", "added2", "key2"}},
		}),
		Entry("should remove key to all tags", testCase{
			tags:         envoy.TagKeysSlice{{"key", "key2"}, {"key", "key3"}},
			transformers: []envoy.TagKeyTransformer{envoy.Without("key")},
			expected:     envoy.TagKeysSlice{envoy.TagKeys{"key2"}, envoy.TagKeys{"key3"}},
		}),
		Entry("should remove dupe after adding key to all tags", testCase{
			tags:         envoy.TagKeysSlice{{"key", "key2"}, {"key"}},
			transformers: []envoy.TagKeyTransformer{envoy.With("key2")},
			expected:     envoy.TagKeysSlice{envoy.TagKeys{"key", "key2"}},
		}),
		Entry("should remove dupe after removing key to all tags", testCase{
			tags:         envoy.TagKeysSlice{{"key", "key2"}, {"key"}},
			transformers: []envoy.TagKeyTransformer{envoy.Without("key2")},
			expected:     envoy.TagKeysSlice{envoy.TagKeys{"key"}},
		}),
		Entry("add then remove different keys works", testCase{
			tags:         envoy.TagKeysSlice{{"key", "key2"}, {"key3"}},
			transformers: []envoy.TagKeyTransformer{envoy.With("key4"), envoy.Without("key")},
			expected:     envoy.TagKeysSlice{envoy.TagKeys{"key2", "key4"}, envoy.TagKeys{"key3", "key4"}},
		}),
		Entry("add then remove same key noop", testCase{
			tags:         envoy.TagKeysSlice{{"key", "key2"}, {"key3"}},
			transformers: []envoy.TagKeyTransformer{envoy.With("key4"), envoy.Without("key4")},
			expected:     envoy.TagKeysSlice{envoy.TagKeys{"key", "key2"}, envoy.TagKeys{"key3"}}.Transform(),
		}),
		Entry("remove then add same key adds it", testCase{
			tags:         envoy.TagKeysSlice{{"key", "key2"}, {"key3"}},
			transformers: []envoy.TagKeyTransformer{envoy.Without("key3"), envoy.With("key3")},
			expected:     envoy.TagKeysSlice{envoy.TagKeys{"key", "key2", "key3"}, envoy.TagKeys{"key3"}}.Transform(),
		}),
	)
})
