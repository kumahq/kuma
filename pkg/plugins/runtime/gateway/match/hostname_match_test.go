package match_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
)

var _ = Describe("Hostname matching", func() {
	DescribeTable("Matches should match",
		func(hostname, candidate string) {
			Expect(match.Hostnames(hostname, candidate)).To(BeTrue())
		},
		Entry("equal exact", "foo.example.com", "foo.example.com"),
		Entry("wild matches exact", "*.example.com", "foo.example.com"),
		Entry("exact matches wild", "foo.example.com", "*.example.com"),
		Entry("wild exact", "*.example.com", "*.example.com"),
		Entry("larger wild matches smaller wild", "*.com", "*.examples.com"),
		Entry("smaller wild matches larger wild", "*.examples.com", "*.com"),
		Entry("larger exact matches smaller wild", "foo.example.com", "*.com"),
		Entry("smaller wild matches larger exact", "foo.examples.com", "*.com"),
	)
	DescribeTable("Matches should not match",
		func(hostname, candidate string) {
			Expect(match.Hostnames(hostname, candidate)).To(BeFalse())
		},
		Entry("exact unequal", "foo.example.com", "bar.example.com"),
		Entry("first wild with rest unequal", "*.example.com", "foo.examples.com"),
		Entry("second wild with rest unequal", "foo.example.com", "*.examples.com"),
		Entry("both wild with rest unequal", "*.example.com", "*.examples.com"),
	)
	DescribeTable("Contains",
		func(hostname, candidate string, expect bool) {
			Expect(match.Contains(hostname, candidate)).To(Equal(expect))
		},
		Entry("equal exact succeeds", "foo.example.com", "foo.example.com", true),
		Entry("exact unequal fails", "foo.example.com", "bar.example.com", false),
		Entry("first wild with rest unequal fails", "*.foo.com", "foo.bar.com", false),
		Entry("first wild with rest equal succeeds", "*.bar.com", "foo.bar.com", true),
		Entry("second wild with rest equal fails", "foo.bar.com", "*.bar.com", false),
		Entry("both wildcards succeeds", "*", "*", true),
		Entry("larger wildcard contains smaller", "*.com", "*.foo.com", true),
		Entry("smaller wildcard doesn't contain larger", "*.foo.com", "*.com", false),
	)
	DescribeTable("SortHostnamesDecByInclusion",
		func(hostnames, expect []string) {
			Expect(match.SortHostnamesByExactnessDec(hostnames)).To(Equal(expect))
		},
		Entry("simple exact", []string{"foo.example.com", "bar.example.com"}, []string{"bar.example.com", "foo.example.com"}),
		Entry("simple exact reverse", []string{"bar.example.com", "foo.example.com"}, []string{"bar.example.com", "foo.example.com"}),
		Entry("exact less than wild", []string{"foo.example.com", "*.example.com"}, []string{"foo.example.com", "*.example.com"}),
		Entry("exact less than wild", []string{"*.example.com", "foo.example.com"}, []string{"foo.example.com", "*.example.com"}),
		Entry("specific wild less than wild", []string{"*.example.com", "*.com"}, []string{"*.example.com", "*.com"}),
		Entry("specific wild less than wild", []string{"*.com", "*.example.com"}, []string{"*.example.com", "*.com"}),
	)
	DescribeTable("SortHostnamesOn",
		func(hostnames, expect []string) {
			Expect(match.SortHostnamesOn(hostnames, func(e string) string { return e })).To(Equal(expect))
		},
		Entry("simple exact", []string{"foo.example.com", "bar.example.com"}, []string{"bar.example.com", "foo.example.com"}),
		Entry("simple exact reverse", []string{"bar.example.com", "foo.example.com"}, []string{"bar.example.com", "foo.example.com"}),
		Entry("exact less than wild", []string{"foo.example.com", "*.example.com"}, []string{"foo.example.com", "*.example.com"}),
		Entry("exact less than wild", []string{"*.example.com", "foo.example.com"}, []string{"foo.example.com", "*.example.com"}),
		Entry("specific wild less than wild", []string{"*.example.com", "*.com"}, []string{"*.example.com", "*.com"}),
		Entry("specific wild less than wild", []string{"*.com", "*.example.com"}, []string{"*.example.com", "*.com"}),
	)
})
