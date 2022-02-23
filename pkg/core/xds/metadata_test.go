package xds_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/matchers"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type testCase struct {
	node     *structpb.Struct
	expected xds.DataplaneMetadata
}

var _ = Describe("DataplaneMetadataFromXdsMetadata", func() {
	DescribeTable("should parse metadata",
		func(given testCase) {
			// when
			metadata := xds.DataplaneMetadataFromXdsMetadata(given.node)

			// then
			Expect(*metadata).To(Equal(given.expected))
		},
		Entry("from empty node", testCase{
			node:     &structpb.Struct{},
			expected: xds.DataplaneMetadata{},
		}),
		Entry("from non-empty node", testCase{
			node: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"dataplane.admin.port": {
						Kind: &structpb.Value_StringValue{
							StringValue: "1234",
						},
					},
					"dataplane.dns.port": {
						Kind: &structpb.Value_StringValue{
							StringValue: "8000",
						},
					},
					"dataplane.dns.empty.port": {
						Kind: &structpb.Value_StringValue{
							StringValue: "8001",
						},
					},
				},
			},
			expected: xds.DataplaneMetadata{
				AdminPort:    1234,
				DNSPort:      8000,
				EmptyDNSPort: 8001,
			},
		}),
	)

	It("should parse version", func() { // this has to be separate test because Equal does not work on proto
		// given
		version := &mesh_proto.Version{
			KumaDp: &mesh_proto.KumaDpVersion{
				Version:   "0.0.1",
				GitTag:    "v0.0.1",
				GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
				BuildDate: "2019-08-07T11:26:06Z",
			},
			Envoy: &mesh_proto.EnvoyVersion{
				Version: "1.15.0",
				Build:   "hash/1.15.0/RELEASE",
			},
		}

		node := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"version": {
					Kind: &structpb.Value_StructValue{
						StructValue: util_proto.MustToStruct(version),
					},
				},
			},
		}

		// when
		metadata := xds.DataplaneMetadataFromXdsMetadata(node)

		// then
		Expect(metadata.Version).To(matchers.MatchProto(version))
	})
})
