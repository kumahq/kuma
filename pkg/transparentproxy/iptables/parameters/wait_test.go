package parameters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

var _ = Describe("WaitParameter", func() {
	DescribeTable("should return",
		func(seconds int, verbose bool, want string) {
			// given
			wait := Wait(uint(seconds))

			// when
			got := wait.Build(verbose)

			// then
			Expect(got).To(Equal(want))
		},
		Entry("no flag when parameter is 0", 0, false, ""),
		Entry("no flag when parameter is 0 - verbose", 0, true, ""),
		Entry("wait seconds", 10, false, "--wait=10"),
		Entry("wait seconds - verbose", 10, true, "--wait=10"),
	)
})
