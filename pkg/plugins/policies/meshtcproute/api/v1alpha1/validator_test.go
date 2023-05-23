package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/core/validators"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/test/resources"
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
  rules:
  - default:
      backendRefs:
      - kind: MeshService
        name: other
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
    kind: BlahBlah
    name: backend
  rules:
  - default:
      backendRefs:
      - kind: MeshService
        name: other
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
		Entry("accepts valid resource", `
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
	)
})
