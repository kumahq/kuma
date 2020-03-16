package xds_test

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/core/xds"
)

type testCase struct {
	node     envoy_core.Node
	expected xds.DataplaneMetadata
}

var _ = DescribeTable("DataplaneMetadataFromNode",
	func(given testCase) {
		// when
		metadata := xds.DataplaneMetadataFromNode(&given.node)

		// then
		Expect(*metadata).To(Equal(given.expected))
	},
	Entry("should parse metadata from empty node", testCase{
		node:     envoy_core.Node{},
		expected: xds.DataplaneMetadata{},
	}),
	Entry("should parse metadata", testCase{
		node: envoy_core.Node{
			Metadata: &pstruct.Struct{
				Fields: map[string]*pstruct.Value{
					"dataplaneTokenPath": &pstruct.Value{
						Kind: &pstruct.Value_StringValue{
							StringValue: "/tmp/token",
						},
					},
					"dataplane.admin.port": &pstruct.Value{
						Kind: &pstruct.Value_StringValue{
							StringValue: "1234",
						},
					},
				},
			},
		},
		expected: xds.DataplaneMetadata{
			DataplaneTokenPath: "/tmp/token",
			AdminPort:          1234,
		},
	}),
)
