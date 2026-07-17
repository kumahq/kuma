package inspect_test

import (
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
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

var (
	meta1 = &test_model.ResourceMeta{Name: "meta1"}
	meta2 = &test_model.ResourceMeta{Name: "meta2"}
	meta3 = &test_model.ResourceMeta{Name: "meta3"}
	meta4 = &test_model.ResourceMeta{Name: "meta4"}
	meta5 = &test_model.ResourceMeta{Name: "meta5"}
	meta6 = &test_model.ResourceMeta{Name: "meta6"}
)

var _ = Describe("GroupByAttachment", func() {
	type testCase struct {
		matchedPolicies *core_xds.MatchedPolicies
		dpNetworking    *mesh_proto.Dataplane_Networking
		expected        inspect.AttachmentMap
	}

	DescribeTable("should generate attachmentMap based on MatchedPolicies",
		func(given testCase) {
			actual := inspect.GroupByAttachment(given.matchedPolicies, given.dpNetworking)
			for k := range actual {
				Expect(actual[k]).To(Equal(given.expected[k]), fmt.Sprintf("attachement %+v", k))
			}
			Expect(actual).To(Equal(given.expected))
		},
		Entry("group by inbounds", testCase{
			matchedPolicies: &core_xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
					core_mesh.HealthCheckType: {
						InboundPolicies: map[mesh_proto.InboundInterface][]core_model.Resource{
							inbound("192.168.0.1", 80, 81): {
								&core_mesh.HealthCheckResource{Meta: meta1},
							},
							inbound("192.168.0.2", 80, 81): {
								&core_mesh.HealthCheckResource{Meta: meta2},
							},
							inbound("192.168.0.2", 90, 91): {
								&core_mesh.HealthCheckResource{Meta: meta3},
							},
						},
					},
					core_mesh.CircuitBreakerType: {
						InboundPolicies: map[mesh_proto.InboundInterface][]core_model.Resource{
							inbound("192.168.0.2", 90, 91): {
								&core_mesh.CircuitBreakerResource{Meta: meta4},
							},
						},
					},
					core_mesh.RateLimitType: {
						InboundPolicies: map[mesh_proto.InboundInterface][]core_model.Resource{
							inbound("192.168.0.2", 90, 91): {
								&core_mesh.RateLimitResource{Meta: meta5},
							},
						},
					},
				},
			},
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
					{
						Address:     "192.168.0.2",
						Port:        90,
						ServicePort: 91,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "web-admin",
						},
					},
				},
			},
			expected: inspect.AttachmentMap{
				inspect.Attachment{Type: inspect.Inbound, Name: "192.168.0.1:80:81", Service: "web"}: {
					core_mesh.HealthCheckType: []core_model.Resource{
						&core_mesh.HealthCheckResource{Meta: meta1},
					},
				},
				inspect.Attachment{Type: inspect.Inbound, Name: "192.168.0.2:80:81", Service: "web-api"}: {
					core_mesh.HealthCheckType: []core_model.Resource{
						&core_mesh.HealthCheckResource{Meta: meta2},
					},
				},
				inspect.Attachment{Type: inspect.Inbound, Name: "192.168.0.2:90:91", Service: "web-admin"}: {
					core_mesh.HealthCheckType: []core_model.Resource{
						&core_mesh.HealthCheckResource{Meta: meta3},
					},
					core_mesh.RateLimitType: []core_model.Resource{
						&core_mesh.RateLimitResource{Meta: meta5},
					},
					core_mesh.CircuitBreakerType: []core_model.Resource{
						&core_mesh.CircuitBreakerResource{Meta: meta4},
					},
				},
			},
		}),
		Entry("group by outbounds", testCase{
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
						Port:    80,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "postgres",
						},
					},
					{
						Address: "192.168.0.2",
						Port:    90,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "mysql",
						},
					},
					{
						Address: "192.168.0.3",
						Port:    90,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "elastic",
						},
					},
					{
						Address: "192.168.0.4",
						Port:    90,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "cockroachdb",
						},
					},
				},
			},
			matchedPolicies: &core_xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
					core_mesh.HealthCheckType: {
						OutboundPolicies: map[mesh_proto.OutboundInterface][]core_model.Resource{
							outbound("192.168.0.1", 80): {
								&core_mesh.HealthCheckResource{Meta: meta1},
							},
							outbound("192.168.0.2", 80): {
								&core_mesh.HealthCheckResource{Meta: meta2},
							},
							outbound("192.168.0.2", 90): {
								&core_mesh.HealthCheckResource{Meta: meta3},
							},
							outbound("192.168.0.4", 90): {
								&core_mesh.HealthCheckResource{Meta: meta5},
							},
						},
					},
					core_mesh.CircuitBreakerType: {
						OutboundPolicies: map[mesh_proto.OutboundInterface][]core_model.Resource{
							outbound("192.168.0.4", 90): {
								&core_mesh.CircuitBreakerResource{Meta: meta6},
							},
						},
					},
					core_mesh.RateLimitType: {
						OutboundPolicies: map[mesh_proto.OutboundInterface][]core_model.Resource{
							outbound("192.168.0.4", 90): {
								&core_mesh.RateLimitResource{Meta: meta6},
							},
						},
					},
				},
			},
			expected: inspect.AttachmentMap{
				inspect.Attachment{Type: inspect.Outbound, Name: "192.168.0.1:80", Service: "redis"}: {
					core_mesh.HealthCheckType: []core_model.Resource{
						&core_mesh.HealthCheckResource{Meta: meta1},
					},
				},
				inspect.Attachment{Type: inspect.Outbound, Name: "192.168.0.2:80", Service: "postgres"}: {
					core_mesh.HealthCheckType: []core_model.Resource{
						&core_mesh.HealthCheckResource{Meta: meta2},
					},
				},
				inspect.Attachment{Type: inspect.Outbound, Name: "192.168.0.2:90", Service: "mysql"}: {
					core_mesh.HealthCheckType: []core_model.Resource{
						&core_mesh.HealthCheckResource{Meta: meta3},
					},
				},
				inspect.Attachment{Type: inspect.Outbound, Name: "192.168.0.4:90", Service: "cockroachdb"}: {
					core_mesh.RateLimitType: []core_model.Resource{
						&core_mesh.RateLimitResource{Meta: meta6},
					},
					core_mesh.CircuitBreakerType: []core_model.Resource{
						&core_mesh.CircuitBreakerResource{Meta: meta6},
					},
					core_mesh.HealthCheckType: []core_model.Resource{
						&core_mesh.HealthCheckResource{Meta: meta5},
					},
				},
			},
		}),
		Entry("group by service", testCase{
			dpNetworking: &mesh_proto.Dataplane_Networking{},
			matchedPolicies: &core_xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
					core_mesh.TrafficLogType: {
						ServicePolicies: map[core_xds.ServiceName][]core_model.Resource{
							"redis": {
								&core_mesh.TrafficLogResource{Meta: meta6},
							},
						},
					},
				},
			},
			expected: inspect.AttachmentMap{
				inspect.Attachment{Type: inspect.Service, Name: "redis", Service: "redis"}: {
					core_mesh.TrafficLogType: []core_model.Resource{
						&core_mesh.TrafficLogResource{Meta: meta6},
					},
				},
			},
		}),
		Entry("group by dataplane", testCase{
			matchedPolicies: &core_xds.MatchedPolicies{
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
					core_mesh.TrafficTraceType: {
						DataplanePolicies: []core_model.Resource{
							&core_mesh.TrafficTraceResource{Meta: meta3},
						},
					},
				},
			},
			expected: inspect.AttachmentMap{
				inspect.Attachment{Type: inspect.Dataplane, Name: ""}: {
					core_mesh.TrafficTraceType: []core_model.Resource{
						&core_mesh.TrafficTraceResource{Meta: meta3},
					},
				},
			},
		}),
	)
})

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
					core_mesh.HealthCheckType: {
						InboundPolicies: map[mesh_proto.InboundInterface][]core_model.Resource{
							inbound("192.168.0.1", 80, 81): {
								&core_mesh.HealthCheckResource{
									Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "default"},
								},
							},
							inbound("192.168.0.2", 90, 91): {
								&core_mesh.HealthCheckResource{
									Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "default"},
								},
							},
							inbound("192.168.0.3", 80, 81): {
								&core_mesh.HealthCheckResource{
									Meta: &test_model.ResourceMeta{Name: "t-2", Mesh: "default"},
								},
							},
						},
					},
				},
			},
			expected: inspect.AttachmentsByPolicy{
				inspect.PolicyKey{
					Type: core_mesh.HealthCheckType,
					Key:  core_model.ResourceKey{Name: "t-1", Mesh: "default"},
				}: {
					{Type: inspect.Inbound, Name: "192.168.0.1:80:81", Service: "web"},
					{Type: inspect.Inbound, Name: "192.168.0.2:90:91", Service: "web-api"},
				},
				inspect.PolicyKey{
					Type: core_mesh.HealthCheckType,
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
					core_mesh.HealthCheckType: {
						OutboundPolicies: map[mesh_proto.OutboundInterface][]core_model.Resource{
							outbound("192.168.0.1", 80): {
								&core_mesh.HealthCheckResource{
									Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "mesh-1"},
								},
							},
							outbound("192.168.0.2", 90): {
								&core_mesh.HealthCheckResource{
									Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "mesh-1"},
								},
							},
						},
					},
				},
			},
			expected: inspect.AttachmentsByPolicy{
				inspect.PolicyKey{
					Type: core_mesh.HealthCheckType,
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
					core_mesh.RateLimitType: {
						InboundPolicies: map[mesh_proto.InboundInterface][]core_model.Resource{
							inbound("192.168.0.1", 80, 81): {
								&core_mesh.RateLimitResource{
									Meta: &test_model.ResourceMeta{Name: "rl-3", Mesh: "mesh-1"},
								},
							},
						},
						OutboundPolicies: map[mesh_proto.OutboundInterface][]core_model.Resource{
							outbound("192.168.0.3", 80): {
								&core_mesh.RateLimitResource{
									Meta: &test_model.ResourceMeta{Name: "rl-3", Mesh: "mesh-1"},
								},
							},
						},
					},
				},
			},
			expected: inspect.AttachmentsByPolicy{
				inspect.PolicyKey{
					Type: core_mesh.RateLimitType,
					Key:  core_model.ResourceKey{Name: "rl-3", Mesh: "mesh-1"},
				}: {
					{Type: inspect.Inbound, Name: "192.168.0.1:80:81", Service: "web"},
					{Type: inspect.Outbound, Name: "192.168.0.3:80", Service: "redis"},
				},
			},
		}),
	)
})
