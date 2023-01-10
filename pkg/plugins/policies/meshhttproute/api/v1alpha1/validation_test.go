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
		ErrorCase("basic error",
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
