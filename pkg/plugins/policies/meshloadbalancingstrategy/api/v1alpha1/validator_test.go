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
			"lb type errors",
			[]validators.Violation{
				{
					Field:   "spec.to[1].default.loadBalancer.leastRequest",
					Message: "must be defined",
				},
				{
					Field:   "spec.to[2].default.loadBalancer.ringHash",
					Message: "must be defined",
				},
				{
					Field:   "spec.to[4].default.loadBalancer.maglev",
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
        type: RoundRobin
  - targetRef:
      kind: MeshService
      name: svc-3
    default:
      loadBalancer:
        type: LeastRequest
  - targetRef:
      kind: MeshService
      name: svc-4
    default:
      loadBalancer:
        type: RingHash
  - targetRef:
      kind: MeshService
      name: svc-5
    default:
      loadBalancer:
        type: Random
  - targetRef:
      kind: MeshService
      name: svc-6
    default:
      loadBalancer:
        type: Maglev`),
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
`))

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
	)
})
