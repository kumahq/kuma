package parameters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters/match/conntrack"
)

var _ = Describe("ConntrackParameter", func() {
	Describe("Ctstate", func() {
		DescribeTable("should build valid ctstate parameter with the combined, "+
			"provided ...State",
			func(states []State, verbose bool, want string) {
				// when
				got := Ctstate(states[0], states[1:]...).Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			// INVALID
			Entry("INVALID state",
				[]State{INVALID}, false,
				"--ctstate INVALID",
			),
			Entry("INVALID state - verbose",
				[]State{INVALID}, true,
				"--ctstate INVALID",
			),
			// NEW
			Entry("NEW state",
				[]State{NEW}, false,
				"--ctstate NEW",
			),
			Entry("NEW state - verbose",
				[]State{NEW}, true,
				"--ctstate NEW",
			),
			// ESTABLISHED
			Entry("ESTABLISHED state",
				[]State{ESTABLISHED}, false,
				"--ctstate ESTABLISHED",
			),
			Entry("ESTABLISHED state - verbose",
				[]State{ESTABLISHED}, true,
				"--ctstate ESTABLISHED",
			),
			// RELATED
			Entry("RELATED state",
				[]State{RELATED}, false,
				"--ctstate RELATED",
			),
			Entry("RELATED state - verbose",
				[]State{RELATED}, true,
				"--ctstate RELATED",
			),
			// UNTRACKED
			Entry("UNTRACKED state",
				[]State{UNTRACKED}, false,
				"--ctstate UNTRACKED",
			),
			Entry("UNTRACKED state - verbose",
				[]State{UNTRACKED}, true,
				"--ctstate UNTRACKED",
			),
			// SNAT
			Entry("SNAT state",
				[]State{SNAT}, false,
				"--ctstate SNAT",
			),
			Entry("SNAT state - verbose",
				[]State{SNAT}, true,
				"--ctstate SNAT",
			),
			// DNAT
			Entry("DNAT state",
				[]State{DNAT}, false,
				"--ctstate DNAT",
			),
			Entry("DNAT state - verbose",
				[]State{DNAT}, true,
				"--ctstate DNAT",
			),
			// Multiple states
			Entry("Multiple states",
				[]State{INVALID, NEW, ESTABLISHED, RELATED, UNTRACKED, SNAT, DNAT}, false,
				"--ctstate INVALID,NEW,ESTABLISHED,RELATED,UNTRACKED,SNAT,DNAT",
			),
			Entry("Multiple states - verbose",
				[]State{INVALID, NEW, ESTABLISHED, RELATED, UNTRACKED, SNAT, DNAT}, true,
				"--ctstate INVALID,NEW,ESTABLISHED,RELATED,UNTRACKED,SNAT,DNAT",
			),
		)

		DescribeTable("should build valid ctstate parameter with the combined, "+
			"provided ...State when negated",
			func(states []State, verbose bool, want string) {
				// when
				got := Ctstate(states[0], states[1:]...).Negate().Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			// INVALID
			Entry("INVALID state",
				[]State{INVALID}, false,
				"! --ctstate INVALID",
			),
			Entry("INVALID state - verbose",
				[]State{INVALID}, true,
				"! --ctstate INVALID",
			),
			// NEW
			Entry("NEW state",
				[]State{NEW}, false,
				"! --ctstate NEW",
			),
			Entry("NEW state - verbose",
				[]State{NEW}, true,
				"! --ctstate NEW",
			),
			// ESTABLISHED
			Entry("ESTABLISHED state",
				[]State{ESTABLISHED}, false,
				"! --ctstate ESTABLISHED",
			),
			Entry("ESTABLISHED state - verbose",
				[]State{ESTABLISHED}, true,
				"! --ctstate ESTABLISHED",
			),
			// RELATED
			Entry("RELATED state",
				[]State{RELATED}, false,
				"! --ctstate RELATED",
			),
			Entry("RELATED state - verbose",
				[]State{RELATED}, true,
				"! --ctstate RELATED",
			),
			// UNTRACKED
			Entry("UNTRACKED state",
				[]State{UNTRACKED}, false,
				"! --ctstate UNTRACKED",
			),
			Entry("UNTRACKED state - verbose",
				[]State{UNTRACKED}, true,
				"! --ctstate UNTRACKED",
			),
			// SNAT
			Entry("SNAT state",
				[]State{SNAT}, false,
				"! --ctstate SNAT",
			),
			Entry("SNAT state - verbose",
				[]State{SNAT}, true,
				"! --ctstate SNAT",
			),
			// DNAT
			Entry("DNAT state",
				[]State{DNAT}, false,
				"! --ctstate DNAT",
			),
			Entry("DNAT state - verbose",
				[]State{DNAT}, true,
				"! --ctstate DNAT",
			),
			// Multiple states
			Entry("Multiple states",
				[]State{INVALID, NEW, ESTABLISHED, RELATED, UNTRACKED, SNAT, DNAT}, false,
				"! --ctstate INVALID,NEW,ESTABLISHED,RELATED,UNTRACKED,SNAT,DNAT",
			),
			Entry("Multiple states - verbose",
				[]State{INVALID, NEW, ESTABLISHED, RELATED, UNTRACKED, SNAT, DNAT}, true,
				"! --ctstate INVALID,NEW,ESTABLISHED,RELATED,UNTRACKED,SNAT,DNAT",
			),
		)

		DescribeTable("Conntrack",
			func(parameters []*ConntrackParameter, verbose bool, want string) {
				// when
				got := Conntrack(parameters...).Build(verbose)

				// then
				Expect(got).To(Equal(want))
			},
			Entry("no parameters",
				nil, false,
				"conntrack",
			),
			Entry("no parameters - verbose",
				nil, true,
				"conntrack",
			),
			Entry("1 parameter (Ctstate with all possible states)",
				[]*ConntrackParameter{Ctstate(
					INVALID, NEW, ESTABLISHED, RELATED, UNTRACKED, SNAT, DNAT,
				)}, false,
				"conntrack --ctstate INVALID,NEW,ESTABLISHED,RELATED,UNTRACKED,SNAT,DNAT",
			),
			Entry("1 parameter (Ctstate with all possible states) - verbose",
				[]*ConntrackParameter{Ctstate(
					INVALID, NEW, ESTABLISHED, RELATED, UNTRACKED, SNAT, DNAT,
				)}, true,
				"conntrack --ctstate INVALID,NEW,ESTABLISHED,RELATED,UNTRACKED,SNAT,DNAT",
			),
		)
	})
})
