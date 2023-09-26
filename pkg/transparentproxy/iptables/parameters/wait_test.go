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
		Entry("wait without argument", 0, false, "-w"),
		Entry("wait without argument- verbose", 0, true, "--wait"),
		Entry("wait seconds", 10, false, "-w 10"),
		Entry("wait seconds - verbose", 10, true, "--wait 10"),
	)
})
