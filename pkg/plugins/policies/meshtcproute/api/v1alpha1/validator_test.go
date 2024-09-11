package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/core/validators"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/test/resources/validators"
)

var _ = Describe("validator", func() {
	DescribeErrorCases(
		api.NewMeshTCPRouteResource,
		ErrorCase("spec.targetRef error",
			validators.Violation{
				Field:   "spec.targetRef.kind",
				Message: "value is not supported",
			}, `
type: MeshTCPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: BlahBlah
  name: frontend
to:
- targetRef:
    kind: MeshService
    name: backend
`),
		ErrorCase("spec.to.targetRef error",
			validators.Violation{
				Field:   "spec.to[0].targetRef.kind",
				Message: "value is not supported",
			}, `
type: MeshTCPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: frontend
to:
- targetRef:
    kind: Mesh
`),
		ErrorCase("spec.to.targetRef error",
			validators.Violation{
				Field:   "spec.to[0].targetRef.kind",
				Message: "value is not supported",
			}, `
type: MeshTCPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshGateway
  name: edge
to:
- targetRef:
    kind: MeshService
    name: backend
`),
		ErrorCase("invalid backendRefs",
			validators.Violation{
				Field:   "spec.to[0].rules[0].default.backendRefs[0].name",
				Message: "must be set with kind MeshServiceSubset",
			}, `
type: MeshTCPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: frontend
to:
- targetRef:
    kind: MeshService
    name: backend
  rules:
  - default:
      backendRefs:
      - kind: MeshServiceSubset
        tags:
          version: v1
`),
	)
	DescribeValidCases(
		api.NewMeshTCPRouteResource,
		Entry("accepts valid resource with to.rules", `
type: MeshTCPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: frontend
to:
- targetRef:
    kind: MeshService
    name: backend
  rules:
  - default:
      backendRefs:
      - kind: MeshService
        name: other
`),
		Entry("accepts valid resource without to.rules", `
type: MeshTCPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: frontend
to:
- targetRef:
    kind: MeshService
    name: backend
`),
		Entry("accepts MeshGateway with listener tags targeted route", `
type: MeshTCPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshGateway
  name: edge
  tags:
    port: 6000
to:
- targetRef:
    kind: Mesh
  rules:
  - default:
      backendRefs:
      - kind: MeshService
        name: other
`),
		Entry("MeshService and MeshMultiZoneService", `
type: MeshTCPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: Mesh
to:
- targetRef:
    kind: MeshService
    name: backend
    sectionName: "8080"
  rules:
  - default:
      backendRefs:
      - kind: MeshMultiZoneService
        name: other
        port: 8080
- targetRef:
    kind: MeshMultiZoneService
    name: other
    sectionName: "8080"
  rules:
  - default:
      backendRefs:
      - kind: MeshService
        name: backend
        port: 8080
`),
	)
})
