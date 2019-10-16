package xds_test

import (
	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

type testCase struct {
	node     core.Node
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
		node:     core.Node{},
		expected: xds.DataplaneMetadata{},
	}),
	Entry("should parse metadata", testCase{
		node: core.Node{
			Metadata: &types.Struct{
				Fields: map[string]*types.Value{
					"dataplaneTokenPath": &types.Value{
						Kind: &types.Value_StringValue{
							StringValue: "/tmp/token",
						},
					},
				},
			},
		},
		expected: xds.DataplaneMetadata{
			DataplaneTokenPath: "/tmp/token",
		},
	}),
)
