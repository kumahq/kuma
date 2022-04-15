package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

var _ = Describe("Miscellaneous Listener configurers", func() {

	type testCase struct {
		opt      ListenerBuilderOpt
		expected string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener := NewListenerBuilder(envoy.APIV3)

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
			expected: "{}",
		}),
		Entry("noop 2", testCase{
			opt:      AddListenerConfigurer(v3.ListenerMustConfigureFunc(nil)),
			expected: "{}",
		}),
		Entry("connection buffer limit", testCase{
			opt:      ConnectionBufferLimit(123),
			expected: "perConnectionBufferLimitBytes: 123",
		}),
		Entry("enable reuse port enabled", testCase{
			opt:      EnableReusePort(true),
			expected: "enableReusePort: true",
		}),
		Entry("enable reuse port disabled", testCase{
			opt:      EnableReusePort(false),
			expected: "enableReusePort: false",
		}),
		Entry("enable freebind", testCase{
			opt:      EnableFreebind(true),
			expected: "freebind: true",
		}),
		Entry("disable freebind", testCase{
			opt:      EnableFreebind(false),
			expected: "freebind: false",
		}),
	)

})
