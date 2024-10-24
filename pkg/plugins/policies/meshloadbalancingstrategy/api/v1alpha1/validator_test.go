package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/core/validators"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/test/resources/validators"
)

var _ = Describe("validation", func() {
	DescribeErrorCases(
		api.NewMeshLoadBalancingStrategyResource,
		ErrorCases(
			"spec errors",
			[]validators.Violation{
				{
					Field:   "spec.targetRef.kind",
					Message: "value is not supported",
				},
				{
					Field:   "spec.to",
					Message: "needs at least one item",
				},
			},
			`
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshGatewayRoute
  name: route-1
to: []
`),
		ErrorCases(
			"spec.to errors",
			[]validators.Violation{{
				Field:   "spec.to[0].targetRef.kind",
				Message: "value is not supported",
			}, {
				Field:   "spec.to[1].default.localityAwareness.crossZone",
				Message: "must not be set: MeshService traffic is local",
			}},
			`
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: svc-1
to: 
  - targetRef:
      kind: MeshServiceSubset
      name: svc-2
      tags:
        version: v1
  - targetRef:
      kind: MeshService
      name: real-mesh-service
      sectionName: http
    default:
      localityAwareness:
        crossZone: {}
`),
		ErrorCases(
			"ringHash error",
			[]validators.Violation{
				{
					Field:   "spec.to[0].default.loadBalancer.ringHash.hashPolicies[0].header",
					Message: "must be defined",
				},
				{
					Field:   "spec.to[0].default.loadBalancer.ringHash.hashPolicies[1].cookie",
					Message: "must be defined",
				},
				{
					Field:   "spec.to[0].default.loadBalancer.ringHash.hashPolicies[2].connection",
					Message: "must be defined",
				},
				{
					Field:   "spec.to[0].default.loadBalancer.ringHash.hashPolicies[3].queryParameter",
					Message: "must be defined",
				},
				{
					Field:   "spec.to[0].default.loadBalancer.ringHash.hashPolicies[4].filterState",
					Message: "must be defined",
				},
			},
			`
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: svc-1
to: 
  - targetRef:
      kind: MeshService
      name: svc-2
    default:
      loadBalancer:
        type: RingHash
        ringHash:
          hashPolicies:
            - type: Header
            - type: Cookie
            - type: Connection
            - type: QueryParameter
            - type: FilterState
`),
		ErrorCases(
			"ringHash cookie error",
			[]validators.Violation{{
				Field:   "spec.to[0].default.loadBalancer.ringHash.hashPolicies[0].cookie.path",
				Message: "must be an absolute path",
			}},
			`
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: svc-1
to: 
  - targetRef:
      kind: MeshService
      name: svc-2
    default:
      loadBalancer:
        type: RingHash
        ringHash:
          hashPolicies:
            - type: Cookie
              cookie:
                name: cookie-name
                ttl: 1s
                path: relative-path
`),
		ErrorCases(
			"maglev error",
			[]validators.Violation{
				{
					Field:   "spec.to[0].default.loadBalancer.maglev.hashPolicies[0].header",
					Message: "must be defined",
				},
				{
					Field:   "spec.to[0].default.loadBalancer.maglev.hashPolicies[1].cookie",
					Message: "must be defined",
				},
				{
					Field:   "spec.to[0].default.loadBalancer.maglev.hashPolicies[3].queryParameter",
					Message: "must be defined",
				},
				{
					Field:   "spec.to[0].default.loadBalancer.maglev.hashPolicies[4].filterState",
					Message: "must be defined",
				},
			},
			`
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: svc-1
to: 
  - targetRef:
      kind: MeshService
      name: svc-2
    default:
      loadBalancer:
        type: Maglev
        maglev:
          hashPolicies:
            - type: Header
            - type: Cookie
            - type: SourceIP
            - type: QueryParameter
            - type: FilterState
`),
		ErrorCases(
			"maglev cookie error",
			[]validators.Violation{{
				Field:   "spec.to[0].default.loadBalancer.maglev.hashPolicies[0].cookie.path",
				Message: "must be an absolute path",
			}},
			`
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: svc-1
to: 
  - targetRef:
      kind: MeshService
      name: svc-2
    default:
      loadBalancer:
        type: Maglev
        maglev:
          hashPolicies:
            - type: Cookie
              cookie:
                name: cookie-name
                ttl: 1s
                path: relative-path
`),
		ErrorCases(
			"leastRequest error",
			[]validators.Violation{{
				Field:   "spec.to[0].default.loadBalancer.leastRequest.activeRequestBias",
				Message: "must be greater or equal then: 0",
			}},
			`
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: svc-1
to:
  - targetRef:
      kind: MeshService
      name: svc-2
    default:
      loadBalancer:
        type: LeastRequest
        leastRequest:
          activeRequestBias: -1
`),
		ErrorCases("empty from in failover", []validators.Violation{{
			Field:   "spec.to[0].default.localityAwareness.crossZone.failover[0].from.zones",
			Message: "must not be empty",
		}}, `
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshService
      name: svc-1
    default:
      localityAwareness:
        crossZone:
          failover:
            - from:
                zones: []
              to: 
                type: None
`),
		ErrorCases("incorrect weight", []validators.Violation{
			{
				Field:   "spec.to[0].default.localityAwareness.localZone.affinityTags[0].weight",
				Message: "must be greater than 0",
			},
			{
				Field:   "spec.to[0].default.localityAwareness.localZone.affinityTags[1].key",
				Message: "must not be empty",
			},
		}, `
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshService
      name: svc-1
    default:
      localityAwareness:
        localZone:
          affinityTags:
            - key: k8s/node
              weight: 0
            - key: ""
              weight: 10
`),
		ErrorCases("mixing affinity tags with and without weights", []validators.Violation{{
			Field:   "spec.to[0].default.localityAwareness.localZone.affinityTags",
			Message: "all or none affinity tags should have weight",
		}}, `
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshService
      name: svc-1
    default:
      localityAwareness:
        localZone:
          affinityTags:
            - key: k8s/node
              weight: 10
            - key: k8s/az
`),
		ErrorCases("percentage can't be zero", []validators.Violation{{
			Field:   "spec.to[0].default.localityAwareness.crossZone.failoverThreshold.percentage",
			Message: "must be greater than 0",
		}}, `
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshService
      name: svc-1
    default:
      localityAwareness:
        crossZone:
          failoverThreshold:
            percentage: 0
`),
		XErrorCases("MeshExternalService can be set only with Mesh", []validators.Violation{{
			Field:   "spec.to[0].targetRef.kind",
			Message: "kind MeshExternalService is only allowed with targetRef.kind: Mesh as it is configured on the Zone Egress and shared by all clients in the mesh",
		}}, `
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshSubset
  tags:
    kuma.io/service: test
to:
  - targetRef:
      kind: MeshExternalService
      name: svc-1
    default:
      localityAwareness:
        disabled: true
      loadBalancer:
        type: LeastRequest
`),
		ErrorCases("percentage is not a parseable number", []validators.Violation{{
			Field:   "spec.to[0].default.localityAwareness.crossZone.failoverThreshold.percentage",
			Message: "string must be a valid number",
		}}, `
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshService
      name: svc-1
    default:
      localityAwareness:
        crossZone:
          failoverThreshold:
            percentage: "hello"
`),
		ErrorCases("broken failover rules", []validators.Violation{
			{
				Field:   "spec.to[0].default.localityAwareness.crossZone.failover[0].to.zones",
				Message: "must be empty when type is None",
			},
			{
				Field:   "spec.to[0].default.localityAwareness.crossZone.failover[1].from.zones[1]",
				Message: "must not be empty",
			},
			{
				Field:   "spec.to[0].default.localityAwareness.crossZone.failover[2].to.zones",
				Message: "must be empty when type is Any",
			},
			{
				Field:   "spec.to[0].default.localityAwareness.crossZone.failover[3].to.zones",
				Message: "must not be empty when type is Only",
			},
			{
				Field:   "spec.to[0].default.localityAwareness.crossZone.failover[4].to.zones",
				Message: "must not be empty when type is Only",
			},
			{
				Field:   "spec.to[0].default.localityAwareness.crossZone.failover[5].to.zones",
				Message: "must not be empty when type is AnyExcept",
			},
			{
				Field:   "spec.to[0].default.localityAwareness.crossZone.failover[6].to.zones",
				Message: "must not be empty when type is AnyExcept",
			},
		}, `
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshService
      name: svc-1
    default:
      localityAwareness:
        crossZone:
          failover:
            - from:
                zones: ["zone-1"]
              to: 
                type: None
                zones: ["zone-1"]
            - from:
                zones: ["zone-1", ""]
              to: 
                type: Any
            - to:
                type: Any
                zones: ["zone-1"]
            - to:
                type: Only
                zones: []
            - to:
                type: Only
            - to:
                type: AnyExcept
                zones: []
            - to:
                type: AnyExcept

`),
		ErrorCases(
			"invalid MeshGateway and to MeshService",
			[]validators.Violation{{
				Field:   "spec.to[0].targetRef.kind",
				Message: "value is not supported, only Mesh is allowed if loadBalancer is set",
			}},
			`
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshGateway
  name: edge-gateway
to:
  - targetRef:
      kind: MeshService
      name: svc-1
    default:
      loadBalancer:
        type: LeastRequest
        leastRequest:
          activeRequestBias: "1.3"
`),
	)

	DescribeValidCases(
		api.NewMeshLoadBalancingStrategyResource,
		Entry(
			"full spec",
			`
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: svc-1
to: 
  - targetRef:
      kind: MeshService
      name: svc-2
    default:
      localityAwareness:
        disabled: true
      loadBalancer:
        type: Maglev
        maglev:
          hashPolicies:
            - type: Cookie
              cookie:
                name: cookie-name
                ttl: 1s
                path: /absolute-path
`),
		Entry(
			"full spec leastRequest",
			`
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: svc-1
to:
  - targetRef:
      kind: MeshService
      name: svc-2
    default:
      localityAwareness:
        disabled: true
      loadBalancer:
        type: LeastRequest
        leastRequest:
          activeRequestBias: "1.3"
`),
		Entry(
			"full locality awareness spec",
			`
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: svc-1
to: 
  - targetRef:
      kind: MeshService
      name: svc-2
    default:
      localityAwareness:
        localZone:
          affinityTags: 
            - key: "k8s/node"
            - key: "k8s/az"
        crossZone:
          failover:
            - from:
                zones: ["zone-1"]
              to: 
                type: Only
                zones: ["zone-2"]
            - from:
                zones: ["zone-3"]
              to:
                type: Any
            - from:
                zones: ["zone-4"]
              to:
                type: AnyExcept
                zones: ["zone-1"]
            - to:
                type: None
          failoverThreshold:
            percentage: 70
`),
		Entry(
			"top level MeshGateway",
			`
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshGateway
  name: edge
  tags:
    name: listener-1
to:
  - targetRef:
      kind: MeshService
      name: svc-2
    default:
      localityAwareness:
        disabled: true
`),
		XEntry(
			"to MeshExternalService",
			`
type: MeshLoadBalancingStrategy
mesh: mesh-1
name: route-1
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshExternalService
      name: mes
    default:
      localityAwareness:
        disabled: true
      loadBalancer:
        type: LeastRequest
        leastRequest:
          activeRequestBias: "1.3"
`),
	)
})
