package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
	. "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	v3 "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners/v3"
)

var _ = Describe("Miscellaneous Listener configurers", func() {
	type testCase struct {
		opt      ListenerBuilderOpt
		expected string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener := NewListenerBuilder(envoy.APIV3, "test_listener")

			listener.Configure(given.opt)

			// then
			resource, err := listener.Build()
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(resource)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("noop 1", testCase{
			opt:      AddListenerConfigurer(v3.ListenerConfigureFunc(nil)),
			expected: "name: test_listener",
		}),
	)
})
