package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/core/validators"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources"
)

var _ = Describe("validation", func() {
	resources.DescribeErrorCases(
		api.NewMeshLoadBalancingStrategyResource,
		resources.ErrorCases(
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
		resources.ErrorCases(
			"spec.to errors",
			[]validators.Violation{{
				Field:   "spec.to[0].targetRef.kind",
				Message: "value is not supported",
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
`),
		resources.ErrorCases(
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
		resources.ErrorCases(
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
		resources.ErrorCases(
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
		resources.ErrorCases(
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
		resources.ErrorCases("", []validators.Violation{{
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
		resources.ErrorCases("", []validators.Violation{{
			Field:   "spec.to[0].default.localityAwareness.localZone.affinityTags[0].weight",
			Message: "must be greater than 1",
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
              weight: -10
`),
		resources.ErrorCases("", []validators.Violation{{
			Field:   "spec.to[0].default.localityAwareness.localZone.affinityTags",
			Message: "each or none affinity tags should have weight",
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
		resources.ErrorCases("", []validators.Violation{{
			Field:   "spec.to[0].default.localityAwareness.failoverThreshold.percentage",
			Message: "has to be in [0.0 - 100.0] range",
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
        failoverThreshold:
          percentage: 200
`),
		resources.ErrorCases("", []validators.Violation{
			{
				Field:   "spec.to[0].default.localityAwareness.crossZone.failover[0].to.zones",
				Message: "must be empty when type is None",
			},
			{
				Field:   "spec.to[0].default.localityAwareness.crossZone.failover[1].to.zones",
				Message: "must be empty when type is Any",
			},
			{
				Field:   "spec.to[0].default.localityAwareness.crossZone.failover[2].to.zones",
				Message: "must not be empty when type is Only",
			},
			{
				Field:   "spec.to[0].default.localityAwareness.crossZone.failover[3].to.zones",
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
            - to:
                type: Any
                zones: ["zone-1"]
            - to:
                type: Only
                zones: []
            - to:
                type: AnyExcept
                zones: []
`),
	)

	resources.DescribeValidCases(
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
	)
})
