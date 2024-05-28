package meshexternalservice_test

import (
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("MeshExternalServiceGenerator", func() {
	type testCase struct {
		ctx        xds_context.Context
		proxy      *core_xds.Proxy
		expected   string
		identity   bool
		usedCas    map[string]struct{}
		allInOneCa bool
	}

	DescribeTable("",
		func(given testCase) {
		},
		Entry("", testCase{}),
	)
})
