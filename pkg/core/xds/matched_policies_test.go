package xds_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

func inbound(ip string, dpPort, workloadPort uint32) mesh_proto.InboundInterface {
	return mesh_proto.InboundInterface{
		DataplaneAdvertisedIP: ip,
		DataplaneIP:           ip,
		DataplanePort:         dpPort,
		WorkloadIP:            ip,
		WorkloadPort:          workloadPort,
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
		expected        core_xds.AttachmentMap
	}

	DescribeTable("should generate attachmentMap based on MatchedPolicies",
		func(given testCase) {
			actual := core_xds.GroupByAttachment(given.matchedPolicies, given.dpNetworking)
			for k := range actual {
				Expect(actual[k]).To(Equal(given.expected[k]), fmt.Sprintf("attachement %+v", k))
			}
			Expect(actual).To(Equal(given.expected))
		},
		Entry("group by inbounds", testCase{
			matchedPolicies: &core_xds.MatchedPolicies{
				TrafficPermissions: core_xds.TrafficPermissionMap{
					inbound("192.168.0.1", 80, 81): {Meta: meta1},
					inbound("192.168.0.2", 80, 81): {Meta: meta2},
					inbound("192.168.0.2", 90, 91): {Meta: meta3},
				},
				FaultInjections: core_xds.FaultInjectionMap{
					inbound("192.168.0.1", 80, 81): {
						{Meta: meta1},
						{Meta: meta4},
					},
					inbound("192.168.0.2", 80, 81): {
						{Meta: meta2},
					},
				},
				RateLimitsInbound: core_xds.InboundRateLimitsMap{
					inbound("192.168.0.2", 90, 91): {
						{Meta: meta3},
					},
				},
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
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
			expected: core_xds.AttachmentMap{
				core_xds.Attachment{Type: core_xds.Inbound, Name: "192.168.0.1:80:81", Service: "web"}: {
					core_mesh.FaultInjectionType: []core_model.Resource{
						&core_mesh.FaultInjectionResource{Meta: meta1},
						&core_mesh.FaultInjectionResource{Meta: meta4},
					},
					core_mesh.TrafficPermissionType: []core_model.Resource{
						&core_mesh.TrafficPermissionResource{Meta: meta1},
					},
				},
				core_xds.Attachment{Type: core_xds.Inbound, Name: "192.168.0.2:80:81", Service: "web-api"}: {
					core_mesh.FaultInjectionType: []core_model.Resource{
						&core_mesh.FaultInjectionResource{Meta: meta2},
					},
					core_mesh.TrafficPermissionType: []core_model.Resource{
						&core_mesh.TrafficPermissionResource{Meta: meta2},
					},
				},
				core_xds.Attachment{Type: core_xds.Inbound, Name: "192.168.0.2:90:91", Service: "web-admin"}: {
					core_mesh.TrafficPermissionType: []core_model.Resource{
						&core_mesh.TrafficPermissionResource{Meta: meta3},
					},
					core_mesh.RateLimitType: []core_model.Resource{
						&core_mesh.RateLimitResource{Meta: meta3},
						&core_mesh.RateLimitResource{Meta: meta5},
					},
					core_mesh.CircuitBreakerType: []core_model.Resource{
						&core_mesh.CircuitBreakerResource{Meta: meta4},
					},
				},
			}}),
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
				Timeouts: core_xds.TimeoutMap{
					outbound("192.168.0.1", 80): {Meta: meta1},
					outbound("192.168.0.2", 80): {Meta: meta2},
					outbound("192.168.0.2", 90): {Meta: meta3},
					outbound("192.168.0.3", 90): {Meta: meta4},
				},
				RateLimitsOutbound: core_xds.OutboundRateLimitsMap{
					outbound("192.168.0.1", 80): {Meta: meta1},
					outbound("192.168.0.2", 80): {Meta: meta2},
					outbound("192.168.0.2", 90): {Meta: meta3},
					outbound("192.168.0.4", 90): {Meta: meta5},
				},
				TrafficRoutes: core_xds.RouteMap{
					outbound("192.168.0.1", 80): {Meta: meta1},
					outbound("192.168.0.2", 80): {Meta: meta2},
					outbound("192.168.0.2", 90): {Meta: meta3},
					outbound("192.168.0.4", 90): {Meta: meta5},
				},
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
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
			expected: core_xds.AttachmentMap{
				core_xds.Attachment{Type: core_xds.Outbound, Name: "192.168.0.1:80", Service: "redis"}: {
					core_mesh.TimeoutType: []core_model.Resource{
						&core_mesh.TimeoutResource{Meta: meta1},
					},
					core_mesh.RateLimitType: []core_model.Resource{
						&core_mesh.RateLimitResource{Meta: meta1},
					},
					core_mesh.TrafficRouteType: []core_model.Resource{
						&core_mesh.TrafficRouteResource{Meta: meta1},
					},
				},
				core_xds.Attachment{Type: core_xds.Outbound, Name: "192.168.0.2:80", Service: "postgres"}: {
					core_mesh.TimeoutType: []core_model.Resource{
						&core_mesh.TimeoutResource{Meta: meta2},
					},
					core_mesh.RateLimitType: []core_model.Resource{
						&core_mesh.RateLimitResource{Meta: meta2},
					},
					core_mesh.TrafficRouteType: []core_model.Resource{
						&core_mesh.TrafficRouteResource{Meta: meta2},
					},
				},
				core_xds.Attachment{Type: core_xds.Outbound, Name: "192.168.0.2:90", Service: "mysql"}: {
					core_mesh.TimeoutType: []core_model.Resource{
						&core_mesh.TimeoutResource{Meta: meta3},
					},
					core_mesh.RateLimitType: []core_model.Resource{
						&core_mesh.RateLimitResource{Meta: meta3},
					},
					core_mesh.TrafficRouteType: []core_model.Resource{
						&core_mesh.TrafficRouteResource{Meta: meta3},
					},
				},
				core_xds.Attachment{Type: core_xds.Outbound, Name: "192.168.0.3:90", Service: "elastic"}: {
					core_mesh.TimeoutType: []core_model.Resource{
						&core_mesh.TimeoutResource{Meta: meta4},
					},
				},
				core_xds.Attachment{Type: core_xds.Outbound, Name: "192.168.0.4:90", Service: "cockroachdb"}: {
					core_mesh.RateLimitType: []core_model.Resource{
						&core_mesh.RateLimitResource{Meta: meta5},
						&core_mesh.RateLimitResource{Meta: meta6},
					},
					core_mesh.CircuitBreakerType: []core_model.Resource{
						&core_mesh.CircuitBreakerResource{Meta: meta6},
					},
					core_mesh.TrafficRouteType: []core_model.Resource{
						&core_mesh.TrafficRouteResource{Meta: meta5},
					},
				},
			},
		}),
		Entry("group by service", testCase{
			dpNetworking: &mesh_proto.Dataplane_Networking{},
			matchedPolicies: &core_xds.MatchedPolicies{
				TrafficLogs: core_xds.TrafficLogMap{
					"backend":  &core_mesh.TrafficLogResource{Meta: meta1},
					"postgres": &core_mesh.TrafficLogResource{Meta: meta2},
				},
				HealthChecks: core_xds.HealthCheckMap{
					"backend": &core_mesh.HealthCheckResource{Meta: meta1},
					"web":     &core_mesh.HealthCheckResource{Meta: meta3},
				},
				CircuitBreakers: core_xds.CircuitBreakerMap{
					"backend":  &core_mesh.CircuitBreakerResource{Meta: meta1},
					"postgres": &core_mesh.CircuitBreakerResource{Meta: meta2},
					"redis":    &core_mesh.CircuitBreakerResource{Meta: meta4},
				},
				Retries: core_xds.RetryMap{
					"backend": &core_mesh.RetryResource{Meta: meta1},
				},
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
			expected: core_xds.AttachmentMap{
				core_xds.Attachment{Type: core_xds.Service, Name: "backend", Service: "backend"}: {
					core_mesh.TrafficLogType: []core_model.Resource{
						&core_mesh.TrafficLogResource{Meta: meta1},
					},
					core_mesh.HealthCheckType: []core_model.Resource{
						&core_mesh.HealthCheckResource{Meta: meta1},
					},
					core_mesh.CircuitBreakerType: []core_model.Resource{
						&core_mesh.CircuitBreakerResource{Meta: meta1},
					},
					core_mesh.RetryType: []core_model.Resource{
						&core_mesh.RetryResource{Meta: meta1},
					},
				},
				core_xds.Attachment{Type: core_xds.Service, Name: "postgres", Service: "postgres"}: {
					core_mesh.TrafficLogType: []core_model.Resource{
						&core_mesh.TrafficLogResource{Meta: meta2},
					},
					core_mesh.CircuitBreakerType: []core_model.Resource{
						&core_mesh.CircuitBreakerResource{Meta: meta2},
					},
				},
				core_xds.Attachment{Type: core_xds.Service, Name: "web", Service: "web"}: {
					core_mesh.HealthCheckType: []core_model.Resource{
						&core_mesh.HealthCheckResource{Meta: meta3},
					},
				},
				core_xds.Attachment{Type: core_xds.Service, Name: "redis", Service: "redis"}: {
					core_mesh.CircuitBreakerType: []core_model.Resource{
						&core_mesh.CircuitBreakerResource{Meta: meta4},
					},
					core_mesh.TrafficLogType: []core_model.Resource{
						&core_mesh.TrafficLogResource{Meta: meta6},
					},
				},
			},
		}),
		Entry("group by dataplane", testCase{
			matchedPolicies: &core_xds.MatchedPolicies{
				TrafficTrace:  &core_mesh.TrafficTraceResource{Meta: meta1},
				ProxyTemplate: &core_mesh.ProxyTemplateResource{Meta: meta2},
				Dynamic: map[core_model.ResourceType]core_xds.TypedMatchingPolicies{
					core_mesh.TrafficTraceType: {
						DataplanePolicies: []core_model.Resource{
							&core_mesh.TrafficTraceResource{Meta: meta3},
						},
					},
				},
			},
			expected: core_xds.AttachmentMap{
				core_xds.Attachment{Type: core_xds.Dataplane, Name: ""}: {
					core_mesh.TrafficTraceType: []core_model.Resource{
						&core_mesh.TrafficTraceResource{Meta: meta1},
						&core_mesh.TrafficTraceResource{Meta: meta3},
					},
					core_mesh.ProxyTemplateType: []core_model.Resource{
						&core_mesh.ProxyTemplateResource{Meta: meta2},
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
		expected        core_xds.AttachmentsByPolicy
	}

	DescribeTable("should generate AttachmentsByPolicy map based on MatchedPolicies",
		func(given testCase) {
			actual := core_xds.GroupByPolicy(given.matchedPolicies, given.dpNetworking)
			for k := range given.expected {
				Expect(actual[k]).To(Equal(given.expected[k]), fmt.Sprintf("policy %+v", k))
			}
		},
		Entry("empty MatchedPolicies", testCase{
			matchedPolicies: &core_xds.MatchedPolicies{},
			expected:        core_xds.AttachmentsByPolicy{},
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
				TrafficPermissions: core_xds.TrafficPermissionMap{
					inbound("192.168.0.1", 80, 81): &core_mesh.TrafficPermissionResource{
						Meta: &test_model.ResourceMeta{Name: "tp-1", Mesh: "default"},
					},
					inbound("192.168.0.2", 90, 91): &core_mesh.TrafficPermissionResource{
						Meta: &test_model.ResourceMeta{Name: "tp-1", Mesh: "default"},
					},
					inbound("192.168.0.3", 80, 81): &core_mesh.TrafficPermissionResource{
						Meta: &test_model.ResourceMeta{Name: "tp-2", Mesh: "default"},
					},
				},
				FaultInjections: core_xds.FaultInjectionMap{
					inbound("192.168.0.1", 80, 81): []*core_mesh.FaultInjectionResource{
						{
							Meta: &test_model.ResourceMeta{Name: "fi-1", Mesh: "default"},
						},
						{
							Meta: &test_model.ResourceMeta{Name: "fi-2", Mesh: "default"},
						},
					},
					inbound("192.168.0.3", 80, 81): []*core_mesh.FaultInjectionResource{
						{
							Meta: &test_model.ResourceMeta{Name: "fi-2", Mesh: "default"},
						},
						{
							Meta: &test_model.ResourceMeta{Name: "fi-3", Mesh: "default"},
						},
					},
				},
				RateLimitsInbound: core_xds.InboundRateLimitsMap{
					inbound("192.168.0.2", 90, 91): []*core_mesh.RateLimitResource{
						{
							Meta: &test_model.ResourceMeta{Name: "rl-1", Mesh: "default"},
						},
					},
				},
			},
			expected: core_xds.AttachmentsByPolicy{
				core_xds.PolicyKey{
					Type: core_mesh.TrafficPermissionType,
					Key:  core_model.ResourceKey{Name: "tp-1", Mesh: "default"},
				}: {
					{Type: core_xds.Inbound, Name: "192.168.0.1:80:81", Service: "web"},
					{Type: core_xds.Inbound, Name: "192.168.0.2:90:91", Service: "web-api"},
				},
				core_xds.PolicyKey{
					Type: core_mesh.TrafficPermissionType,
					Key:  core_model.ResourceKey{Name: "tp-2", Mesh: "default"},
				}: {
					{Type: core_xds.Inbound, Name: "192.168.0.3:80:81", Service: "web-admin"},
				},
				core_xds.PolicyKey{
					Type: core_mesh.FaultInjectionType,
					Key:  core_model.ResourceKey{Name: "fi-1", Mesh: "default"},
				}: {
					{Type: core_xds.Inbound, Name: "192.168.0.1:80:81", Service: "web"},
				},
				core_xds.PolicyKey{
					Type: core_mesh.FaultInjectionType,
					Key:  core_model.ResourceKey{Name: "fi-2", Mesh: "default"},
				}: {
					{Type: core_xds.Inbound, Name: "192.168.0.1:80:81", Service: "web"},
					{Type: core_xds.Inbound, Name: "192.168.0.3:80:81", Service: "web-admin"},
				},
				core_xds.PolicyKey{
					Type: core_mesh.FaultInjectionType,
					Key:  core_model.ResourceKey{Name: "fi-3", Mesh: "default"},
				}: {
					{Type: core_xds.Inbound, Name: "192.168.0.3:80:81", Service: "web-admin"},
				},
				core_xds.PolicyKey{
					Type: core_mesh.RateLimitType,
					Key:  core_model.ResourceKey{Name: "rl-1", Mesh: "default"},
				}: {
					{Type: core_xds.Inbound, Name: "192.168.0.2:90:91", Service: "web-api"},
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
				Timeouts: core_xds.TimeoutMap{
					outbound("192.168.0.1", 80): &core_mesh.TimeoutResource{
						Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "mesh-1"},
					},
					outbound("192.168.0.2", 90): &core_mesh.TimeoutResource{
						Meta: &test_model.ResourceMeta{Name: "t-1", Mesh: "mesh-1"},
					},
				},
				RateLimitsOutbound: core_xds.OutboundRateLimitsMap{
					outbound("192.168.0.1", 80): &core_mesh.RateLimitResource{
						Meta: &test_model.ResourceMeta{Name: "rl-1", Mesh: "mesh-1"},
					},
					outbound("192.168.0.2", 90): &core_mesh.RateLimitResource{
						Meta: &test_model.ResourceMeta{Name: "rl-2", Mesh: "mesh-1"},
					},
				},
			},
			expected: core_xds.AttachmentsByPolicy{
				core_xds.PolicyKey{
					Type: core_mesh.TimeoutType,
					Key:  core_model.ResourceKey{Name: "t-1", Mesh: "mesh-1"},
				}: {
					{Type: core_xds.Outbound, Name: "192.168.0.1:80", Service: "redis"},
					{Type: core_xds.Outbound, Name: "192.168.0.2:90", Service: "postgres"},
				},
				core_xds.PolicyKey{
					Type: core_mesh.RateLimitType,
					Key:  core_model.ResourceKey{Name: "rl-1", Mesh: "mesh-1"},
				}: {
					{Type: core_xds.Outbound, Name: "192.168.0.1:80", Service: "redis"},
				},
				core_xds.PolicyKey{
					Type: core_mesh.RateLimitType,
					Key:  core_model.ResourceKey{Name: "rl-2", Mesh: "mesh-1"},
				}: {
					{Type: core_xds.Outbound, Name: "192.168.0.2:90", Service: "postgres"},
				},
			},
		}),
		Entry("group by service policies", testCase{
			dpNetworking: &mesh_proto.Dataplane_Networking{},
			matchedPolicies: &core_xds.MatchedPolicies{
				TrafficLogs: core_xds.TrafficLogMap{
					"backend": &core_mesh.TrafficLogResource{
						Meta: &test_model.ResourceMeta{Name: "tl-1", Mesh: "mesh-1"},
					},
					"postgres": &core_mesh.TrafficLogResource{
						Meta: &test_model.ResourceMeta{Name: "tl-1", Mesh: "mesh-1"},
					},
					"redis": &core_mesh.TrafficLogResource{
						Meta: &test_model.ResourceMeta{Name: "tl-2", Mesh: "mesh-1"},
					},
				},
				HealthChecks: core_xds.HealthCheckMap{
					"backend": &core_mesh.HealthCheckResource{
						Meta: &test_model.ResourceMeta{Name: "hc-1", Mesh: "mesh-1"},
					},
					"redis": &core_mesh.HealthCheckResource{
						Meta: &test_model.ResourceMeta{Name: "hc-2", Mesh: "mesh-1"},
					},
				},
				CircuitBreakers: core_xds.CircuitBreakerMap{
					"kafka": &core_mesh.CircuitBreakerResource{
						Meta: &test_model.ResourceMeta{Name: "cb-1", Mesh: "mesh-1"},
					},
				},
				Retries: core_xds.RetryMap{
					"payments": &core_mesh.RetryResource{
						Meta: &test_model.ResourceMeta{Name: "r-1", Mesh: "mesh-1"},
					},
					"backend": &core_mesh.RetryResource{
						Meta: &test_model.ResourceMeta{Name: "r-2", Mesh: "mesh-1"},
					},
					"web": &core_mesh.RetryResource{
						Meta: &test_model.ResourceMeta{Name: "r-2", Mesh: "mesh-1"},
					},
				},
			},
			expected: core_xds.AttachmentsByPolicy{
				core_xds.PolicyKey{
					Type: core_mesh.TrafficLogType,
					Key:  core_model.ResourceKey{Name: "tl-1", Mesh: "mesh-1"},
				}: {
					{Type: core_xds.Service, Name: "backend", Service: "backend"},
					{Type: core_xds.Service, Name: "postgres", Service: "postgres"},
				},
				core_xds.PolicyKey{
					Type: core_mesh.TrafficLogType,
					Key:  core_model.ResourceKey{Name: "tl-2", Mesh: "mesh-1"},
				}: {
					{Type: core_xds.Service, Name: "redis", Service: "redis"},
				},
				core_xds.PolicyKey{
					Type: core_mesh.HealthCheckType,
					Key:  core_model.ResourceKey{Name: "hc-1", Mesh: "mesh-1"},
				}: {
					{Type: core_xds.Service, Name: "backend", Service: "backend"},
				},
				core_xds.PolicyKey{
					Type: core_mesh.HealthCheckType,
					Key:  core_model.ResourceKey{Name: "hc-2", Mesh: "mesh-1"},
				}: {
					{Type: core_xds.Service, Name: "redis", Service: "redis"},
				},
				core_xds.PolicyKey{
					Type: core_mesh.CircuitBreakerType,
					Key:  core_model.ResourceKey{Name: "cb-1", Mesh: "mesh-1"},
				}: {
					{Type: core_xds.Service, Name: "kafka", Service: "kafka"},
				},
				core_xds.PolicyKey{
					Type: core_mesh.RetryType,
					Key:  core_model.ResourceKey{Name: "r-1", Mesh: "mesh-1"},
				}: {
					{Type: core_xds.Service, Name: "payments", Service: "payments"},
				},
				core_xds.PolicyKey{
					Type: core_mesh.RetryType,
					Key:  core_model.ResourceKey{Name: "r-2", Mesh: "mesh-1"},
				}: {
					{Type: core_xds.Service, Name: "backend", Service: "backend"},
					{Type: core_xds.Service, Name: "web", Service: "web"},
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
				RateLimitsOutbound: core_xds.OutboundRateLimitsMap{
					outbound("192.168.0.3", 80): &core_mesh.RateLimitResource{
						Meta: &test_model.ResourceMeta{Name: "rl-1", Mesh: "mesh-1"},
					},
					outbound("192.168.0.4", 80): &core_mesh.RateLimitResource{
						Meta: &test_model.ResourceMeta{Name: "rl-1", Mesh: "mesh-1"},
					},
				},
				RateLimitsInbound: core_xds.InboundRateLimitsMap{
					inbound("192.168.0.1", 80, 81): []*core_mesh.RateLimitResource{
						{
							Meta: &test_model.ResourceMeta{Name: "rl-1", Mesh: "mesh-1"},
						},
					},
					inbound("192.168.0.2", 80, 81): []*core_mesh.RateLimitResource{
						{
							Meta: &test_model.ResourceMeta{Name: "rl-1", Mesh: "mesh-1"},
						},
					},
				},
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
			expected: core_xds.AttachmentsByPolicy{
				core_xds.PolicyKey{
					Type: core_mesh.RateLimitType,
					Key:  core_model.ResourceKey{Name: "rl-1", Mesh: "mesh-1"},
				}: {
					{Type: core_xds.Inbound, Name: "192.168.0.1:80:81", Service: "web"},
					{Type: core_xds.Inbound, Name: "192.168.0.2:80:81", Service: "web-api"},
					{Type: core_xds.Outbound, Name: "192.168.0.3:80", Service: "redis"},
					{Type: core_xds.Outbound, Name: "192.168.0.4:80", Service: "postgres"},
				},
				core_xds.PolicyKey{
					Type: core_mesh.RateLimitType,
					Key:  core_model.ResourceKey{Name: "rl-3", Mesh: "mesh-1"},
				}: {
					{Type: core_xds.Inbound, Name: "192.168.0.1:80:81", Service: "web"},
					{Type: core_xds.Outbound, Name: "192.168.0.3:80", Service: "redis"},
				},
			},
		}),
	)
})
