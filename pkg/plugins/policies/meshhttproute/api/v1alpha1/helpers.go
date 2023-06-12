package v1alpha1

type PolicyDefault struct {
	Rules []Rule `policyMerge:"mergeValuesByKey"`
}

func (x *To) GetDefault() interface{} {
	var rules []Rule

	for _, rule := range x.Rules {
		var matches []Match

		for _, match := range rule.Matches {
			if match.Path == nil {
				// According to Envoy docs, match must have precisely one of
				// prefix, path, safe_regex, connect_matcher,
				// path_separated_prefix, path_match_policy set, so when policy
				// doesn't specify explicit type of matching, we are assuming
				// "catch all" path (any path starting with "/").
				match.Path = &PathMatch{
					Value: "/",
					Type:  PathPrefix,
				}
			}

			matches = append(matches, match)
		}

		rule.Matches = matches

		rules = append(rules, rule)
	}

	return PolicyDefault{
		Rules: rules,
	}
}
