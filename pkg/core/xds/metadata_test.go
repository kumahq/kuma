package xds_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
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
			metadata := xds.DataplaneMetadataFromXdsMetadata(given.node, "/tmp", core_model.ResourceKey{
				Name: "dp-1",
				Mesh: "mesh",
			})

			// then
			Expect(*metadata).To(Equal(given.expected))
		},
		Entry("from empty node", testCase{
			node: &structpb.Struct{},
			expected: xds.DataplaneMetadata{
				AccessLogSocketPath: "/tmp/kuma-al-dp-1-mesh.sock",
				MetricsSocketPath:   "/tmp/kuma-mh-dp-1-mesh.sock",
			},
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
					"dataplane.readinessReporter.port": {
						Kind: &structpb.Value_StringValue{
							StringValue: "9300",
						},
					},
					"accessLogSocketPath": {
						Kind: &structpb.Value_StringValue{
							StringValue: "/tmp/logs",
						},
					},
					"metricsSocketPath": {
						Kind: &structpb.Value_StringValue{
							StringValue: "/tmp/metrics",
						},
					},
				},
			},
			expected: xds.DataplaneMetadata{
				AdminPort:           1234,
				DNSPort:             8000,
				EmptyDNSPort:        8001,
				AccessLogSocketPath: "/tmp/logs",
				MetricsSocketPath:   "/tmp/metrics",
				ReadinessPort:       9300,
			},
		}),
		Entry("should ignore dependencies version provided through metadata if version is not set at all", testCase{
			node: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"dynamicMetadata": {
						Kind: &structpb.Value_StructValue{
							StructValue: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"version.dependencies.coredns": {
										Kind: &structpb.Value_StringValue{
											StringValue: "8000",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: xds.DataplaneMetadata{
				AccessLogSocketPath: "/tmp/kuma-al-dp-1-mesh.sock",
				MetricsSocketPath:   "/tmp/kuma-mh-dp-1-mesh.sock",
				DynamicMetadata:     map[string]string{},
			},
		}),
	)

	It("should fallback to service side generated paths", func() { // remove with https://github.com/kumahq/kuma/issues/7220
		// given
		dpJSON, err := json.Marshal(rest.From.Resource(&core_mesh.DataplaneResource{
			Meta: &test_model.ResourceMeta{Mesh: "mesh", Name: "dp-1"},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "123.40.2.2",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{Address: "10.0.0.1", Port: 8080, Tags: map[string]string{"kuma.io/service": "foo"}},
					},
				},
			},
		}))
		Expect(err).ToNot(HaveOccurred())
		node := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"dataplane.resource": {
					Kind: &structpb.Value_StringValue{
						StringValue: string(dpJSON),
					},
				},
			},
		}

		// when
		metadata := xds.DataplaneMetadataFromXdsMetadata(node, "/tmp", core_model.ResourceKey{
			Name: "dp-1",
			Mesh: "mesh",
		})

		// then
		Expect(metadata.AccessLogSocketPath).To(Equal("/tmp/kuma-al-dp-1-mesh.sock"))
		Expect(metadata.MetricsSocketPath).To(Equal("/tmp/kuma-mh-dp-1-mesh.sock"))
	})

	It("should fallback to service side generated paths without dpp in metadata", func() { // remove with https://github.com/kumahq/kuma/issues/7220
		// given
		node := &structpb.Struct{
			Fields: map[string]*structpb.Value{},
		}

		// when
		metadata := xds.DataplaneMetadataFromXdsMetadata(node, "/tmp", core_model.ResourceKey{
			Name: "dp-1",
			Mesh: "mesh",
		})

		// then
		Expect(metadata.AccessLogSocketPath).To(Equal("/tmp/kuma-al-dp-1-mesh.sock"))
		Expect(metadata.MetricsSocketPath).To(Equal("/tmp/kuma-mh-dp-1-mesh.sock"))
	})

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
		metadata := xds.DataplaneMetadataFromXdsMetadata(node, "/tmp", core_model.ResourceKey{
			Name: "dp-1",
			Mesh: "mesh",
		})

		// then
		// We don't want to validate KumaDpVersion.KumaCpCompatible
		// as compatibility checks are currently checked in insights
		// ref: https://github.com/kumahq/kuma/issues/4203
		Expect(metadata.GetVersion().GetEnvoy()).
			To(matchers.MatchProto(version.GetEnvoy()))
		Expect(metadata.GetVersion().GetKumaDp().GetVersion()).
			To(Equal(version.GetKumaDp().GetVersion()))
		Expect(metadata.GetVersion().GetKumaDp().GetBuildDate()).
			To(Equal(version.GetKumaDp().GetBuildDate()))
		Expect(metadata.GetVersion().GetKumaDp().GetGitCommit()).
			To(Equal(version.GetKumaDp().GetGitCommit()))
		Expect(metadata.GetVersion().GetKumaDp().GetGitTag()).
			To(Equal(version.GetKumaDp().GetGitTag()))
	})
})
