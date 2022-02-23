package dns_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/dns"
)

var _ = Describe("DNS name modifier", func() {
	type testCase struct {
		input    string
		expected string
	}

	DescribeTable("should modify names",
		func(given testCase) {
			// when
			actual, err := dns.DnsNameToKumaCompliant(given.input)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("empty", testCase{
			input:    "",
			expected: "",
		}),
		Entry("one last dot", testCase{
			input:    "mesh.",
			expected: "mesh.",
		}),
		Entry("one dot", testCase{
			input:    ".mesh",
			expected: ".mesh",
		}),
		Entry("two dots with last", testCase{
			input:    "a.mesh.",
			expected: "a.mesh.",
		}),
		Entry("two dots", testCase{
			input:    "a.b.mesh",
			expected: "a_b.mesh",
		}),
		Entry("three dots with last", testCase{
			input:    "a.b.mesh.",
			expected: "a_b.mesh.",
		}),
		Entry("three dots", testCase{
			input:    "a.b.c.mesh",
			expected: "a_b_c.mesh",
		}),
	)

})
