package matchers

import (
	xds_config "github.com/cncf/xds/go/xds/core/v3"
	matcher_config "github.com/cncf/xds/go/xds/type/matcher/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	rbac_config "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	matcher_extension "github.com/envoyproxy/go-control-plane/envoy/extensions/common/matching/v3"
	actionv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/matcher/action/v3"
	sslv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/matching/common_inputs/ssl/v3"
	"google.golang.org/protobuf/types/known/anypb"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	util_slices "github.com/kumahq/kuma/pkg/util/slices"
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

func FieldMatcher(fieldMatcherBuilder *Builder[matcher_config.Matcher_MatcherList_FieldMatcher]) Configurer[matcher_config.Matcher] {
	fieldMatcher, err := fieldMatcherBuilder.Build()
	if err != nil {
		return nil
	}
	return MatchersList([]*matcher_config.Matcher_MatcherList_FieldMatcher{fieldMatcher})
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

func NewFieldMatcher() *Builder[matcher_config.Matcher_MatcherList_FieldMatcher] {
	return &Builder[matcher_config.Matcher_MatcherList_FieldMatcher]{}
}

func Matches(matches []common_api.Match, onMatchBuilder *Builder[matcher_config.Matcher_OnMatch]) Configurer[matcher_config.Matcher_MatcherList_FieldMatcher] {
	return matchesConfigurer(matches, onMatchBuilder, false)
}

func NotMatches(matches []common_api.Match, onMatchBuilder *Builder[matcher_config.Matcher_OnMatch]) Configurer[matcher_config.Matcher_MatcherList_FieldMatcher] {
	return matchesConfigurer(matches, onMatchBuilder, true)
}

func matchesConfigurer(matches []common_api.Match, onMatchBuilder *Builder[matcher_config.Matcher_OnMatch], negate bool) Configurer[matcher_config.Matcher_MatcherList_FieldMatcher] {
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
			if match.SpiffeID != nil {
				predicate, err := NewPredicate().Configure(SpiffeIDPredicate(match.SpiffeID)).Build()
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

		if negate {
			combinedPredicate, err = NewPredicate().Configure(NotPredicate(combinedPredicate)).Build()
			if err != nil {
				return err
			}
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

func SkipFilterAction() Configurer[matcher_config.Matcher_OnMatch] {
	return func(onMatch *matcher_config.Matcher_OnMatch) error {
		onMatch.OnMatch = &matcher_config.Matcher_OnMatch_Action{
			Action: &xds_config.TypedExtensionConfig{
				Name:        "skip",
				TypedConfig: util_proto.MustMarshalAny(&actionv3.SkipFilter{}),
			},
		}
		return nil
	}
}

func NewPredicate() *Builder[matcher_config.Matcher_MatcherList_Predicate] {
	return &Builder[matcher_config.Matcher_MatcherList_Predicate]{}
}

func SpiffeIDPredicate(spiffeID *common_api.SpiffeIDMatch) Configurer[matcher_config.Matcher_MatcherList_Predicate] {
	return func(predicate *matcher_config.Matcher_MatcherList_Predicate) error {
		var stringMatcher matcher_config.StringMatcher
		switch spiffeID.Type {
		case common_api.ExactMatchType:
			stringMatcher = matcher_config.StringMatcher{
				MatchPattern: &matcher_config.StringMatcher_Exact{
					Exact: spiffeID.Value,
				},
			}
		case common_api.PrefixMatchType:
			stringMatcher = matcher_config.StringMatcher{
				MatchPattern: &matcher_config.StringMatcher_Prefix{
					Prefix: spiffeID.Value,
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

func NotPredicate(predicate *matcher_config.Matcher_MatcherList_Predicate) Configurer[matcher_config.Matcher_MatcherList_Predicate] {
	return func(notPredicate *matcher_config.Matcher_MatcherList_Predicate) error {
		notPredicate.MatchType = &matcher_config.Matcher_MatcherList_Predicate_NotMatcher{
			NotMatcher: predicate,
		}
		return nil
	}
}

func NewExtensionWithMatcher() *Builder[matcher_extension.ExtensionWithMatcher] {
	return &Builder[matcher_extension.ExtensionWithMatcher]{}
}

func Matcher(matcherBuilder *Builder[matcher_config.Matcher]) Configurer[matcher_extension.ExtensionWithMatcher] {
	return func(extension *matcher_extension.ExtensionWithMatcher) error {
		matcher, err := matcherBuilder.Build()
		if err != nil {
			return err
		}
		extension.XdsMatcher = matcher
		return nil
	}
}

func Filter(name string, filter *anypb.Any) Configurer[matcher_extension.ExtensionWithMatcher] {
	return func(extension *matcher_extension.ExtensionWithMatcher) error {
		extension.ExtensionConfig = &corev3.TypedExtensionConfig{
			Name:        name,
			TypedConfig: filter,
		}
		return nil
	}
}
