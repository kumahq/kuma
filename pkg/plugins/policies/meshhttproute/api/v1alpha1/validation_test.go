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
		ErrorCase("invalid filter",
			validators.Violation{
				Field:   `spec.to[0].rules[0].filters[0].requestHeaderModifier`,
				Message: validators.MustBeDefined,
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
    - default:
        filters:
          - type: RequestHeaderModifier
`),
		ErrorCases("prefix rewrite without prefix match",
			[]validators.Violation{{
				Field:   `spec.to[0].rules[0].filters[0].urlRewrite.path.replacePrefixMatch`,
				Message: "can only appear if all matches match a path prefix",
			}, {
				Field:   `spec.to[0].rules[0].filters[1].requestRedirect.path.replacePrefixMatch`,
				Message: "can only appear if all matches match a path prefix",
			}}, `
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
          value: /prefix
          type: Exact
      default:
        filters:
          - type: URLRewrite
            urlRewrite:
              path:
                type: ReplacePrefixMatch
                replacePrefixMatch: /other
          - type: RequestRedirect
            requestRedirect:
              path:
                type: ReplacePrefixMatch
                replacePrefixMatch: /other
`),
		ErrorCases("non-empty value for header present/absent match",
			[]validators.Violation{{
				Field:   `spec.to[0].rules[0].matches[0].headers[0].value`,
				Message: validators.MustNotBeDefined,
			}, {
				Field:   `spec.to[0].rules[0].matches[0].headers[1].value`,
				Message: validators.MustNotBeDefined,
			}}, `
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
      - headers:
        - type: Present
          name: foo
          value: x
        - type: Absent
          name: foo
          value: x
`),
		ErrorCases("invalid backendRef in requestMirror",
			[]validators.Violation{{
				Field:   `spec.to[0].rules[0].filters[0].requestMirror.backendRef.name`,
				Message: "must be set with kind MeshServiceSubset",
			}}, `
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
          type: Prefix
          value: /
      default:
        filters:
          - type: RequestMirror
            requestMirror:
              backendRef:
                kind: MeshServiceSubset
                tags:
                  version: v1
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
		Entry("prefix rewrite with prefix match", `
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
          value: /prefix
          type: Prefix
      default:
        filters:
          - type: URLRewrite
            urlRewrite:
              path:
                type: ReplacePrefixMatch
                replacePrefixMatch: /other
          - type: RequestRedirect
            requestRedirect:
              path:
                type: ReplacePrefixMatch
                replacePrefixMatch: /other
`),
	)
})
