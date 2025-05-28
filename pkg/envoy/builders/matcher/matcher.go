package matcher

import (
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
)

func NewMetadataBuilder() *Builder[matcherv3.MetadataMatcher] {
	return &Builder[matcherv3.MetadataMatcher]{}
}

func Key(filter, key string) Configurer[matcherv3.MetadataMatcher] {
	return func(m *matcherv3.MetadataMatcher) error {
		m.Filter = filter
		m.Path = []*matcherv3.MetadataMatcher_PathSegment{
			{
				Segment: &matcherv3.MetadataMatcher_PathSegment_Key{
					Key: key,
				},
			},
		}
		return nil
	}
}

func ExactValue(value string) Configurer[matcherv3.MetadataMatcher] {
	return func(m *matcherv3.MetadataMatcher) error {
		m.Value = &matcherv3.ValueMatcher{
			MatchPattern: &matcherv3.ValueMatcher_StringMatch{
				StringMatch: &matcherv3.StringMatcher{
					MatchPattern: &matcherv3.StringMatcher_Exact{
						Exact: value,
					},
				},
			},
		}
		return nil
	}
}

func NullValue() Configurer[matcherv3.MetadataMatcher] {
	return func(m *matcherv3.MetadataMatcher) error {
		m.Value = &matcherv3.ValueMatcher{
			MatchPattern: &matcherv3.ValueMatcher_NullMatch_{
				NullMatch: &matcherv3.ValueMatcher_NullMatch{},
			},
		}
		return nil
	}
}
