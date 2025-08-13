package matchers

import (
	xds_config "github.com/cncf/xds/go/xds/core/v3"
	matcher_config "github.com/cncf/xds/go/xds/type/matcher/v3"
	rbac_config "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	sslv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/matching/common_inputs/ssl/v3"
	util_slices "github.com/kumahq/kuma/pkg/util/slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func NewMatcherBuilder() *Builder[matcher_config.Matcher] {
	return &Builder[matcher_config.Matcher]{}
}

func MatchersList(fieldMatchers []*matcher_config.Matcher_MatcherList_FieldMatcher) Configurer[matcher_config.Matcher] {
	return func(matcher *matcher_config.Matcher) error {
		// filter remove empty field matchers
		fieldMatchers = util_slices.Filter(fieldMatchers, func(fm *matcher_config.Matcher_MatcherList_FieldMatcher) bool {
			return fm != nil && fm.Predicate != nil
		})

		if len(fieldMatchers) > 0 {
			matcher.MatcherType = &matcher_config.Matcher_MatcherList_{
				MatcherList: &matcher_config.Matcher_MatcherList{
					Matchers: fieldMatchers,
				},
			}
		}
		return nil
	}
}

func OnNoMatch(actionBuilder *Builder[matcher_config.Matcher_OnMatch]) Configurer[matcher_config.Matcher] {
	return func(matcher *matcher_config.Matcher) error {
		action, err := actionBuilder.Build()
		if err != nil {
			return err
		}
		matcher.OnNoMatch = action
		return nil
	}
}

func NewFieldMatcherList() *Builder[matcher_config.Matcher_MatcherList_FieldMatcher] {
	return &Builder[matcher_config.Matcher_MatcherList_FieldMatcher]{}
}

func Matches(matches []common_api.Match, onMatchBuilder *Builder[matcher_config.Matcher_OnMatch]) Configurer[matcher_config.Matcher_MatcherList_FieldMatcher] {
	return func(fm *matcher_config.Matcher_MatcherList_FieldMatcher) error {
		onMatch, err := onMatchBuilder.Build()
		if err != nil {
			return err
		}

		if len(matches) == 0 {
			return nil
		}

		var predicates []*matcher_config.Matcher_MatcherList_Predicate
		for _, match := range matches {
			switch {
			case match.SpiffeId != nil:
				predicate, err := NewPredicate().Configure(SpiffeIdPredicate(match.SpiffeId)).Build()
				if err != nil {
					return err
				}
				predicates = append(predicates, predicate)
			}
		}

		combinedPredicate, err := NewPredicate().Configure(CombinedPredicate(predicates)).Build()
		if err != nil {
			return err
		}

		fm.Predicate = combinedPredicate
		fm.OnMatch = onMatch
		return nil
	}
}

func NewOnMatch() *Builder[matcher_config.Matcher_OnMatch] {
	return &Builder[matcher_config.Matcher_OnMatch]{}
}

func RbacAction(action rbac_config.RBAC_Action, name string) Configurer[matcher_config.Matcher_OnMatch] {
	return func(onMatch *matcher_config.Matcher_OnMatch) error {
		onMatch.OnMatch = &matcher_config.Matcher_OnMatch_Action{
			Action: &xds_config.TypedExtensionConfig{
				Name: "envoy.filters.rbac.action",
				TypedConfig: util_proto.MustMarshalAny(&rbac_config.Action{
					Name:   name,
					Action: action,
				}),
			},
		}
		return nil
	}
}

func NewPredicate() *Builder[matcher_config.Matcher_MatcherList_Predicate] {
	return &Builder[matcher_config.Matcher_MatcherList_Predicate]{}
}

func SpiffeIdPredicate(spiffeId *common_api.SpiffeIdMatch) Configurer[matcher_config.Matcher_MatcherList_Predicate] {
	return func(predicate *matcher_config.Matcher_MatcherList_Predicate) error {
		var stringMatcher matcher_config.StringMatcher
		switch spiffeId.Type {
		case common_api.ExactMatchType:
			stringMatcher = matcher_config.StringMatcher{
				MatchPattern: &matcher_config.StringMatcher_Exact{
					Exact: spiffeId.Value,
				},
			}
		case common_api.PrefixMatchType:
			stringMatcher = matcher_config.StringMatcher{
				MatchPattern: &matcher_config.StringMatcher_Prefix{
					Prefix: spiffeId.Value,
				},
			}
		}
		predicate.MatchType = &matcher_config.Matcher_MatcherList_Predicate_SinglePredicate_{
			SinglePredicate: &matcher_config.Matcher_MatcherList_Predicate_SinglePredicate{
				Input: &xds_config.TypedExtensionConfig{
					Name:        "envoy.matching.inputs.uri_san",
					TypedConfig: util_proto.MustMarshalAny(&sslv3.UriSanInput{}),
				},
				Matcher: &matcher_config.Matcher_MatcherList_Predicate_SinglePredicate_ValueMatch{
					ValueMatch: &stringMatcher,
				},
			},
		}
		return nil
	}
}

func CombinedPredicate(matchers []*matcher_config.Matcher_MatcherList_Predicate) Configurer[matcher_config.Matcher_MatcherList_Predicate] {
	return func(predicate *matcher_config.Matcher_MatcherList_Predicate) error {
		if len(matchers) == 1 {
			predicate.MatchType = matchers[0].MatchType
		} else if len(matchers) > 1 {
			predicate.MatchType = &matcher_config.Matcher_MatcherList_Predicate_OrMatcher{
				OrMatcher: &matcher_config.Matcher_MatcherList_Predicate_PredicateList{
					Predicate: matchers,
				},
			}
		}
		return nil
	}
}
