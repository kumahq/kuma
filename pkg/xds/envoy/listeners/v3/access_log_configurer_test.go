package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
	. "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners/v3"
)

var _ = Describe("InterpolateKumaValues()", func() {
	format := "%KUMA_SOURCE_ADDRESS% %KUMA_SOURCE_ADDRESS_WITHOUT_PORT% %KUMA_SOURCE_SERVICE%->%KUMA_DESTINATION_SERVICE%"

	DescribeTable("should render source address",
		func(direction envoy.TrafficDirection, expected string) {
			values := KumaValues{
				SourceService:      "backend",
				SourceIP:           "10.0.0.1",
				DestinationService: "web",
				TrafficDirection:   direction,
			}
			Expect(InterpolateKumaValues(format, values)).To(Equal(expected))
		},
		Entry("from the proxy address on outbound",
			envoy.TrafficDirectionOutbound,
			"10.0.0.1:0 10.0.0.1 backend->web",
		),
		Entry("from the proxy address when direction is unspecified",
			envoy.TrafficDirectionUnspecified,
			"10.0.0.1:0 10.0.0.1 backend->web",
		),
		Entry("from the downstream connection on inbound",
			envoy.TrafficDirectionInbound,
			"%DOWNSTREAM_REMOTE_ADDRESS% %DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT% backend->web",
		),
	)
})
