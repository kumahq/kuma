package xds_test

import (
	_struct "github.com/golang/protobuf/ptypes/struct"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	pstruct "google.golang.org/protobuf/types/known/structpb"

	"github.com/kumahq/kuma/pkg/core/xds"
)

type testCase struct {
	node     *_struct.Struct
	expected xds.DataplaneMetadata
}

var _ = DescribeTable("DataplaneMetadataFromXdsMetadata",
	func(given testCase) {
		// when
		metadata := xds.DataplaneMetadataFromXdsMetadata(given.node)

		// then
		Expect(*metadata).To(Equal(given.expected))
	},
	Entry("should parse metadata from empty node", testCase{
		node:     &_struct.Struct{},
		expected: xds.DataplaneMetadata{},
	}),
	Entry("should parse metadata", testCase{
		node: &pstruct.Struct{
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
		expected: xds.DataplaneMetadata{
			DataplaneTokenPath: "/tmp/token",
			AdminPort:          1234,
		},
	}),
)
