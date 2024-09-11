package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func makeSingleMatchRules(matches []api.Match) []api.Rule {
	var rules []api.Rule
	for _, match := range matches {
		rules = append(rules, api.Rule{
			Matches: []api.Match{match},
		})
	}

	return rules
}

func makeMultiMatchRules(matcheses [][]api.Match) []api.Rule {
	var rules []api.Rule
	for _, matches := range matcheses {
		rules = append(rules, api.Rule{
			Matches: matches,
		})
	}

	return rules
}

func makeRoutes(matches []api.Match) []api.Route {
	var routes []api.Route
	for _, match := range matches {
		routes = append(routes, api.Route{
			Hash:  api.HashMatches([]api.Match{match}),
			Match: match,
		})
	}

	return routes
}

var _ = Describe("SortRules", func() {
	exactMatch := api.Match{
		Path: &api.PathMatch{
			Type:  api.Exact,
			Value: "/exact",
		},
	}
	otherExactMatch := api.Match{
		Path: &api.PathMatch{
			Type:  api.Exact,
			Value: "/other-exact",
		},
	}
	prefixMatch := api.Match{
		Path: &api.PathMatch{
			Type:  api.PathPrefix,
			Value: "/prefix",
		},
	}
	longerPrefixMatch := api.Match{
		Path: &api.PathMatch{
			Type:  api.PathPrefix,
			Value: "/prefix/plusmore",
		},
	}
	regexMatch := api.Match{
		Path: &api.PathMatch{
			Type:  api.RegularExpression,
			Value: "/exact.*",
		},
	}
	methodMatch := api.Match{
		Method: pointer.To(api.Method("GET")),
	}
	exactAndMethodMatch := api.Match{
		Path: &api.PathMatch{
			Type:  api.Exact,
			Value: "/exact",
		},
		Method: pointer.To(api.Method("GET")),
	}
	singleHeaderMatch := api.Match{
		Headers: []common_api.HeaderMatch{{
			Type:  pointer.To(common_api.HeaderMatchExact),
			Name:  "header",
			Value: "value",
		}},
	}
	exactSingleHeaderMatch := api.Match{
		Path: &api.PathMatch{
			Type:  api.Exact,
			Value: "/exact",
		},
		Headers: []common_api.HeaderMatch{{
			Type:  pointer.To(common_api.HeaderMatchExact),
			Name:  "header",
			Value: "value",
		}},
	}
	exactDoubleHeaderMatch := api.Match{
		Path: &api.PathMatch{
			Type:  api.Exact,
			Value: "/other-exact",
		},
		Headers: []common_api.HeaderMatch{{
			Type:  pointer.To(common_api.HeaderMatchExact),
			Name:  "header",
			Value: "value",
		}, {
			Type:  pointer.To(common_api.HeaderMatchExact),
			Name:  "other-header",
			Value: "other-value",
		}},
	}
	It("handles base cases", func() {
		Expect(api.SortRules(makeSingleMatchRules([]api.Match{}), nil, nil)).To(Equal(makeRoutes([]api.Match{})))
		Expect(api.SortRules(makeSingleMatchRules([]api.Match{
			exactMatch,
		}), nil, nil)).To(Equal(makeRoutes([]api.Match{
			exactMatch,
		})))
	})
	It("handles path matches", func() {
		Expect(api.SortRules(makeSingleMatchRules([]api.Match{
			prefixMatch,
			longerPrefixMatch,
			regexMatch,
			exactMatch,
		}), nil, nil)).To(Equal(makeRoutes([]api.Match{
			exactMatch,
			longerPrefixMatch,
			prefixMatch,
			regexMatch,
		})))
	})
	It("handles different kinds matches", func() {
		Expect(api.SortRules(makeSingleMatchRules([]api.Match{
			methodMatch,
			exactMatch,
		}), nil, nil)).To(Equal(makeRoutes([]api.Match{
			exactMatch,
			methodMatch,
		})))
		Expect(api.SortRules(makeSingleMatchRules([]api.Match{
			singleHeaderMatch,
			exactMatch,
		}), nil, nil)).To(Equal(makeRoutes([]api.Match{
			exactMatch,
			singleHeaderMatch,
		})))
		Expect(api.SortRules(makeSingleMatchRules([]api.Match{
			singleHeaderMatch,
			prefixMatch,
		}), nil, nil)).To(Equal(makeRoutes([]api.Match{
			prefixMatch,
			singleHeaderMatch,
		})))
	})
	It("handles AND matches", func() {
		Expect(api.SortRules(makeSingleMatchRules([]api.Match{
			exactMatch,
			exactAndMethodMatch,
		}), nil, nil)).To(Equal(makeRoutes([]api.Match{
			exactAndMethodMatch,
			exactMatch,
		})))
		Expect(api.SortRules(makeSingleMatchRules([]api.Match{
			exactSingleHeaderMatch,
			exactDoubleHeaderMatch,
		}), nil, nil)).To(Equal(makeRoutes([]api.Match{
			exactDoubleHeaderMatch,
			exactSingleHeaderMatch,
		})))
	})
	It("is stable", func() {
		Expect(api.SortRules(makeSingleMatchRules([]api.Match{
			exactMatch,
			otherExactMatch,
		}), nil, nil)).To(Equal(makeRoutes([]api.Match{
			exactMatch,
			otherExactMatch,
		})))
		Expect(api.SortRules(makeSingleMatchRules([]api.Match{
			otherExactMatch,
			exactMatch,
		}), nil, nil)).To(Equal(makeRoutes([]api.Match{
			otherExactMatch,
			exactMatch,
		})))
	})
	It("handles rules with multiple matches", func() {
		prefixMatchOnly := []api.Match{prefixMatch}
		singleOrExact := []api.Match{singleHeaderMatch, exactMatch}
		Expect(api.SortRules(makeMultiMatchRules([][]api.Match{
			prefixMatchOnly,
			singleOrExact,
		}), nil, nil)).To(Equal([]api.Route{{
			Match: exactMatch,
			Hash:  api.HashMatches(singleOrExact),
		}, {
			Match: prefixMatch,
			Hash:  api.HashMatches(prefixMatchOnly),
		}, {
			Match: singleHeaderMatch,
			Hash:  api.HashMatches(singleOrExact),
		}}))
	})
})
