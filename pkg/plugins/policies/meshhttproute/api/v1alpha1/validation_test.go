package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/core/validators"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/test/resources"
)

var _ = Describe("validation", func() {
	DescribeErrorCases(
		api.NewMeshHTTPRouteResource,
		ErrorCase("spec.targetRef error",
			validators.Violation{
				Field:   `spec.targetRef.kind`,
				Message: `value is not supported`,
			}, `
type: MeshHTTPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: BlahBlah
  name: frontend
to: []
`),
		ErrorCase("spec.to.targetRef error",
			validators.Violation{
				Field:   `spec.to[0].targetRef.kind`,
				Message: `value is not supported`,
			}, `
type: MeshHTTPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: frontend
to:
- targetRef:
    kind: BlahBlah
    name: frontend
`),
		ErrorCase("empty path match",
			validators.Violation{
				Field:   `spec.to[0].rules[0].matches[0].path.value`,
				Message: `must be an absolute path`,
			}, `
type: MeshHTTPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: frontend
to:
- targetRef:
    kind: MeshService
    name: frontend
  rules:
    - matches:
      - path:
          value: "relative"
          type: Exact
`),
		ErrorCase("repeated match query param names",
			validators.Violation{
				Field:   `spec.to[0].rules[0].matches[0].queryParams[1].name`,
				Message: `multiple entries for name foo`,
			}, `
type: MeshHTTPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: frontend
to:
- targetRef:
    kind: MeshService
    name: frontend
  rules:
    - matches:
      - queryParams:
        - type: Exact
          name: foo
          value: bar
        - type: Exact
          name: foo
          value: baz
`),
	)
	DescribeValidCases(
		api.NewMeshHTTPRouteResource,
		Entry("accepts valid resource", `
type: MeshHTTPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshService
  name: frontend
to: []
`),
	)
})
