package parameters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

var _ = Describe("WaitIntervalParameter", func() {
	DescribeTable("should return",
		func(microseconds int, verbose bool, want string) {
			// given
			waitInterval := WaitInterval(uint(microseconds))

			// when
			got := waitInterval.Build(verbose)

			// then
			Expect(got).To(Equal(want))
		},
		Entry("no flag when parameter is 0", 0, false, ""),
		Entry("no flag when parameter is 0 - verbose", 0, true, ""),
		Entry("wait interval", 20, false, "--wait-interval=20"),
		Entry("wait interval - verbose", 20, true, "--wait-interval=20"),
	)
})
