package v3_test

import (
	envoy_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

var _ = Describe("StatPrefixConfigurer", func() {
	type testCase struct {
		name       string
		address    string
		port       uint32
		protocol   core_xds.SocketAddressProtocol
		statPrefix string
		expected   string
	}

	It("should set statPrefix when provided", func() {
		// given
		listener := &envoy_listener_v3.Listener{}
		configurer := &envoy_listeners_v3.StatPrefixConfigurer{
			StatPrefix: "my-prefix",
		}

		// when
		err := configurer.Configure(listener)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(listener.StatPrefix).To(Equal("my-prefix"))
	})

	It("should not modify statPrefix when empty", func() {
		// given
		listener := &envoy_listener_v3.Listener{
			StatPrefix: "existing",
		}
		configurer := &envoy_listeners_v3.StatPrefixConfigurer{
			StatPrefix: "",
		}

		// when
		err := configurer.Configure(listener)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(listener.StatPrefix).To(Equal("existing"))
	})

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// given
			builder := envoy_listeners.NewOutboundListenerBuilder(envoy_common.APIV3, given.address, given.port, given.protocol).
				WithOverwriteName(given.name)

			// when
			listener, err := builder.
				Configure(envoy_listeners.StatPrefix(given.statPrefix)).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			// and, when
			actual, err := util_proto.ToYAML(listener)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic tcp_proxy", testCase{
			address:    "127.0.0.1",
			port:       5432,
			name:       "my-name",
			statPrefix: "db",
			expected: `
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 5432
            name: my-name
            statPrefix: db
            trafficDirection: OUTBOUND
`,
		}),
	)
})
