package v2_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/xds"

	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("TransparentProxyingConfigurer", func() {

	type testCase struct {
		listenerName        string
		listenerProtocol    xds.SocketAddressProtocol
		listenerAddress     string
		listenerPort        uint32
		transparentProxying *mesh_proto.Dataplane_Networking_TransparentProxying
		expected            string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder(envoy.APIV2).
				Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Configure(TransparentProxying(given.transparentProxying)).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic listener with transparent proxying", testCase{
			listenerName:    "inbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			transparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
				RedirectPortOutbound: 12345,
				RedirectPortInbound:  12346,
			},
			expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            deprecatedV1:
              bindToPort: false
`,
		}),
		Entry("basic listener without transparent proxying", testCase{
			listenerName:        "inbound:192.168.0.1:8080",
			listenerAddress:     "192.168.0.1",
			listenerPort:        8080,
			transparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{},
			expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
`,
		}),
	)
})
