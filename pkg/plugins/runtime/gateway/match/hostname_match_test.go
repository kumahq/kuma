package match_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
)

var _ = Describe("Hostname matching", func() {
	DescribeTable("should match",
		func(hostname string, candidate string) {
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
	DescribeTable("should not match",
		func(hostname string, candidate string) {
			Expect(match.Hostnames(hostname, candidate)).To(BeFalse())
		},
		Entry("exact unequal", "foo.example.com", "bar.example.com"),
		Entry("first wild with rest unequal", "*.example.com", "foo.examples.com"),
		Entry("second wild with rest unequal", "foo.example.com", "*.examples.com"),
		Entry("both wild with rest unequal", "*.example.com", "*.examples.com"),
	)
})
