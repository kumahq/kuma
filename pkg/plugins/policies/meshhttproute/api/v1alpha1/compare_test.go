package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	"slices"
)

func sort(matches []api.Match) []api.Match {
	slices.SortStableFunc(matches, api.CompareMatch)
	return matches
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
		Headers: &[]common_api.HeaderMatch{{
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
		Headers: &[]common_api.HeaderMatch{{
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
		Headers: &[]common_api.HeaderMatch{{
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
		Expect(sort([]api.Match{})).To(Equal([]api.Match{}))
		Expect([]api.Match{exactMatch}).To(Equal([]api.Match{exactMatch}))
	})
	It("handles path matches", func() {
		Expect(sort([]api.Match{
			prefixMatch,
			longerPrefixMatch,
			regexMatch,
			exactMatch,
		})).To(Equal([]api.Match{
			exactMatch,
			longerPrefixMatch,
			prefixMatch,
			regexMatch,
		}))
	})
	It("handles different kinds matches", func() {
		Expect(sort([]api.Match{
			methodMatch,
			exactMatch,
		})).To(Equal([]api.Match{
			exactMatch,
			methodMatch,
		}))
		Expect(sort([]api.Match{
			singleHeaderMatch,
			exactMatch,
		})).To(Equal([]api.Match{
			exactMatch,
			singleHeaderMatch,
		}))
		Expect(sort([]api.Match{
			singleHeaderMatch,
			prefixMatch,
		})).To(Equal([]api.Match{
			prefixMatch,
			singleHeaderMatch,
		}))
	})
	It("handles AND matches", func() {
		Expect(sort([]api.Match{
			exactMatch,
			exactAndMethodMatch,
		})).To(Equal([]api.Match{
			exactAndMethodMatch,
			exactMatch,
		}))
		Expect(sort([]api.Match{
			exactSingleHeaderMatch,
			exactDoubleHeaderMatch,
		})).To(Equal([]api.Match{
			exactDoubleHeaderMatch,
			exactSingleHeaderMatch,
		}))
	})
	It("is stable", func() {
		Expect(sort([]api.Match{
			exactMatch,
			otherExactMatch,
		})).To(Equal([]api.Match{
			exactMatch,
			otherExactMatch,
		}))
		Expect(sort([]api.Match{
			otherExactMatch,
			exactMatch,
		})).To(Equal([]api.Match{
			otherExactMatch,
			exactMatch,
		}))
	})
})
