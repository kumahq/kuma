package gateway_test

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
		Entry("", "foo.example.com", "foo.example.com"),
		Entry("", "*.example.com", "foo.example.com"),
		Entry("", "foo.example.com", "*.example.com"),
		Entry("", "*.example.com", "*.example.com"),
	)
	DescribeTable("should not match",
		func(hostname string, candidate string) {
			Expect(match.Hostnames(hostname, candidate)).To(BeFalse())
		},
		Entry("", "foo.example.com", "bar.example.com"),
		Entry("", "*.example.com", "foo.examples.com"),
		Entry("", "foo.example.com", "*.examples.com"),
		Entry("", "*.example.com", "*.examples.com"),
	)
})
