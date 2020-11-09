package v3_test

import (
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("OutboundListenerConfigurer", func() {

	type testCase struct {
		listenerName    string
		listenerAddress string
		listenerPort    uint32
		expected        string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
<<<<<<< HEAD:pkg/xds/envoy/listeners/v3/outbound_listener_configurer_test.go
			listener, err := NewListenerBuilder(envoy.APIV3).
				Configure(OutboundListener(given.listenerName, given.listenerAddress, given.listenerPort)).
=======
			listener, err := NewListenerBuilder().
				Configure(OutboundListener(given.listenerName, mesh_core.ProtocolTCP, given.listenerAddress, given.listenerPort)).
>>>>>>> fix(*) remove isUDP:pkg/xds/envoy/listeners/outbound_listener_configurer_test.go
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic listener", testCase{
			listenerName:    "outbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			expected: `
            name: outbound:192.168.0.1:8080
            trafficDirection: OUTBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
`,
		}),
	)
})
