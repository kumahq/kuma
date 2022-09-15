package route_test

import (
	"sort"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
)

var _ = Describe("sorting", func() {
	It("is consistent", func() {
		routes := []route.Entry{{
			Match: route.Match{
				ExactPath: "/go",
			},
		}, {
			Match: route.Match{
				PrefixPath: "/go",
			},
		}}

		sort.Sort(route.Sorter(routes))

		swappedRoutes := []route.Entry{{
			Match: route.Match{
				PrefixPath: "/go",
			},
		}, {
			Match: route.Match{
				ExactPath: "/go",
			},
		}}

		sort.Sort(route.Sorter(swappedRoutes))

		Expect(routes).To(Equal(swappedRoutes))
	})
})
