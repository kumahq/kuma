package inspect_test

import (
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/core/xds/inspect"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

func inbound(ip string, dpPort, workloadPort uint32) mesh_proto.InboundInterface {
	return mesh_proto.InboundInterface{
		DataplaneAdvertisedIP: ip,
		DataplaneIP:           ip,
		DataplanePort:         dpPort,
		WorkloadIP:            ip,
		WorkloadPort:          workloadPort,
		InboundName:           strconv.Itoa(int(dpPort)),
	}
}

func outbound(ip string, port uint32) mesh_proto.OutboundInterface {
	return mesh_proto.OutboundInterface{
		DataplaneIP:   ip,
		DataplanePort: port,
	}
}

var _ = Describe("GroupByPolicy", func() {
	type testCase struct {
		matchedPolicies *core_xds.MatchedPolicies
		dpNetworking    *mesh_proto.Dataplane_Networking
		expected        inspect.AttachmentsByPolicy
	}

	DescribeTable("should generate AttachmentsByPolicy map based on MatchedPolicies",
		func(given testCase) {
			actual := inspect.GroupByPolicy(given.matchedPolicies, given.dpNetworking)
			for k := range given.expected {
				Expect(actual[k]).To(Equal(given.expected[k]), fmt.Sprintf("policy %+v", k))
			}
		},
		Entry("empty MatchedPolicies", testCase{
			matchedPolicies: &core_xds.MatchedPolicies{},
			expected:        inspect.AttachmentsByPolicy{},
		}),
		Entry("group by inbound policies", testCase{
			dpNetworking: &mesh_proto.Dataplane_Networking{
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Address:     "192.168.0.1",
						Port:        80,
						ServicePort: 81,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "web",
						},
					},
					{
						Address:     "192.168.0.2",
						Port:        90,
						ServicePort: 91,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "web-api",
						},
					},
					{
						Address:     "192.168.0.3",
						Port:        80,
						ServicePort: 81,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "web-admin",
						},
					},
				},
			},
			matchedPolicies: &core_xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
					meshexternalservice_api.MeshExternalServiceType: {
						InboundPolicies: map[mesh_proto.InboundInterface][]core_model.Resource{
							inbound("192.168.0.1", 80, 81): {
								&meshexternalservice_api.MeshExternalServiceResource{
									Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "default"},
								},
							},
							inbound("192.168.0.2", 90, 91): {
								&meshexternalservice_api.MeshExternalServiceResource{
									Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "default"},
								},
							},
							inbound("192.168.0.3", 80, 81): {
								&meshexternalservice_api.MeshExternalServiceResource{
									Meta: &test_model.ResourceMeta{Name: "t-2", Mesh: "default"},
								},
							},
						},
					},
				},
			},
			expected: inspect.AttachmentsByPolicy{
				inspect.PolicyKey{
					Type: meshexternalservice_api.MeshExternalServiceType,
					Key:  core_model.ResourceKey{Name: "t-1", Mesh: "default"},
				}: {
					{Type: inspect.Inbound, Name: "192.168.0.1:80:81", Service: "web"},
					{Type: inspect.Inbound, Name: "192.168.0.2:90:91", Service: "web-api"},
				},
				inspect.PolicyKey{
					Type: meshexternalservice_api.MeshExternalServiceType,
					Key:  core_model.ResourceKey{Name: "t-2", Mesh: "default"},
				}: {
					{Type: inspect.Inbound, Name: "192.168.0.3:80:81", Service: "web-admin"},
				},
			},
		}),
		Entry("group by outbound policies", testCase{
			dpNetworking: &mesh_proto.Dataplane_Networking{
				Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
					{
						Address: "192.168.0.1",
						Port:    80,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "redis",
						},
					},
					{
						Address: "192.168.0.2",
						Port:    90,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "postgres",
						},
					},
				},
			},
			matchedPolicies: &core_xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
					meshexternalservice_api.MeshExternalServiceType: {
						OutboundPolicies: map[mesh_proto.OutboundInterface][]core_model.Resource{
							outbound("192.168.0.1", 80): {
								&meshexternalservice_api.MeshExternalServiceResource{
									Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "mesh-1"},
								},
							},
							outbound("192.168.0.2", 90): {
								&meshexternalservice_api.MeshExternalServiceResource{
									Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "mesh-1"},
								},
							},
						},
					},
				},
			},
			expected: inspect.AttachmentsByPolicy{
				inspect.PolicyKey{
					Type: meshexternalservice_api.MeshExternalServiceType,
					Key:  core_model.ResourceKey{Name: "t-1", Mesh: "mesh-1"},
				}: {
					{Type: inspect.Outbound, Name: "192.168.0.1:80", Service: "redis"},
					{Type: inspect.Outbound, Name: "192.168.0.2:90", Service: "postgres"},
				},
			},
		}),
		Entry("group by policy that exists both for inbounds and outbounds", testCase{
			dpNetworking: &mesh_proto.Dataplane_Networking{
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Address:     "192.168.0.1",
						Port:        80,
						ServicePort: 81,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "web",
						},
					},
					{
						Address:     "192.168.0.2",
						Port:        80,
						ServicePort: 81,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "web-api",
						},
					},
				},
				Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
					{
						Address: "192.168.0.3",
						Port:    80,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "redis",
						},
					},
					{
						Address: "192.168.0.4",
						Port:    80,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "postgres",
						},
					},
				},
			},
			matchedPolicies: &core_xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
					core_mesh.ZoneIngressType: {
						InboundPolicies: map[mesh_proto.InboundInterface][]core_model.Resource{
							inbound("192.168.0.1", 80, 81): {
								&core_mesh.ZoneIngressResource{
									Meta: &test_model.ResourceMeta{Name: "rl-3", Mesh: "mesh-1"},
								},
							},
						},
						OutboundPolicies: map[mesh_proto.OutboundInterface][]core_model.Resource{
							outbound("192.168.0.3", 80): {
								&core_mesh.ZoneIngressResource{
									Meta: &test_model.ResourceMeta{Name: "rl-3", Mesh: "mesh-1"},
								},
							},
						},
					},
				},
			},
			expected: inspect.AttachmentsByPolicy{
				inspect.PolicyKey{
					Type: core_mesh.ZoneIngressType,
					Key:  core_model.ResourceKey{Name: "rl-3", Mesh: "mesh-1"},
				}: {
					{Type: inspect.Inbound, Name: "192.168.0.1:80:81", Service: "web"},
					{Type: inspect.Outbound, Name: "192.168.0.3:80", Service: "redis"},
				},
			},
		}),
	)
})
