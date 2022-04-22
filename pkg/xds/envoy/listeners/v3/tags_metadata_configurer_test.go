package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("TagsMetadataConfigurer", func() {

	type testCase struct {
		listenerName     string
		listenerProtocol xds.SocketAddressProtocol
		listenerAddress  string
		listenerPort     uint32
		tags             map[string]string
		expected         string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder(envoy_common.APIV3).
				Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Configure(TagsMetadata(given.tags)).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("prometheus endpoint without transparent proxying", testCase{
			listenerName:    "kuma:metrics:prometheus",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			tags: map[string]string{
				"kuma.io/service": "backend",
				"version":         "v2",
			},
			expected: `
            name: kuma:metrics:prometheus
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            enableReusePort: false
            metadata:
              filterMetadata:
                io.kuma.tags:
                  kuma.io/service: backend
                  version: v2
            trafficDirection: INBOUND`,
		}),
	)

})
