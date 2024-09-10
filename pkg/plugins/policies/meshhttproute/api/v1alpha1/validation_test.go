package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/core/validators"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/test/resources/validators"
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
		ErrorCase("spec.to.targetRef Mesh not allowed",
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
    kind: Mesh
`),
		ErrorCase("spec.to.targetRef MeshService not allowed with top MeshGateway",
			validators.Violation{
				Field:   `spec.to[0].targetRef.kind`,
				Message: `value is not supported`,
			}, `
type: MeshHTTPRoute
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
		ErrorCases("incorrect path match value",
			[]validators.Violation{{
				Field:   `spec.to[0].rules[0].matches[0].path.value`,
				Message: `must be an absolute path`,
			}, {
				Field:   `spec.to[0].rules[0].matches[1].path.value`,
				Message: "does not need a trailing slash because only a `/`-separated prefix or an entire path is matched",
			}, {
				Field:   `spec.to[0].rules[0].matches[2].path.value`,
				Message: `must be an absolute path`,
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
          value: "relative"
          type: Exact
      - path:
          value: "/trailing/slash/"
          type: PathPrefix
      - path:
          value: "relative"
          type: PathPrefix
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
		ErrorCases("empty filter",
			[]validators.Violation{{
				Field:   `spec.to[0].rules[0].default.filters[0].requestHeaderModifier`,
				Message: validators.MustBeDefined,
			}, {
				Field:   `spec.to[0].rules[0].default.filters[1].responseHeaderModifier`,
				Message: validators.MustBeDefined,
			}, {
				Field:   `spec.to[0].rules[0].default.filters[2].requestRedirect`,
				Message: validators.MustBeDefined,
			}, {
				Field:   `spec.to[0].rules[0].default.filters[3].requestMirror`,
				Message: validators.MustBeDefined,
			}, {
				Field:   `spec.to[0].rules[0].default.filters[4].urlRewrite`,
				Message: validators.MustBeDefined,
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
    - default:
        filters:
          - type: RequestHeaderModifier
          - type: ResponseHeaderModifier
          - type: RequestRedirect
          - type: RequestMirror
          - type: URLRewrite
`),
		ErrorCases("invalid HeaderModifier filter",
			[]validators.Violation{
				{
					Field:   `spec.to[0].rules[0].default.filters[0].requestHeaderModifier.set[1].name`,
					Message: `duplicate header name`,
				}, {
					Field:   `spec.to[0].rules[0].default.filters[1].responseHeaderModifier`,
					Message: `must have at least one defined: set, add, remove`,
				}, {
					Field:   `spec.to[0].rules[0].default.filters[2].responseHeaderModifier.add[0].name`,
					Message: `duplicate header name`,
				}, {
					Field:   `spec.to[0].rules[0].default.filters[2].responseHeaderModifier.remove[0].name`,
					Message: `duplicate header name`,
				},
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
            requestHeaderModifier:
              set:
                - name: foo
                  value: bar
                - name: foo
                  value: bazz
          - type: ResponseHeaderModifier
            responseHeaderModifier: {}
          - type: ResponseHeaderModifier
            responseHeaderModifier: 
              set:
                - name: foo
                  value: bar
              add:
                - name: foo
                  value: bar
              remove:
                - foo
`),
		ErrorCases("prefix rewrite without prefix match",
			[]validators.Violation{{
				Field:   `spec.to[0].rules[0].default.filters[0].urlRewrite.path.replacePrefixMatch`,
				Message: "can only appear if all matches match a path prefix",
			}, {
				Field:   `spec.to[0].rules[0].default.filters[1].requestRedirect.path.replacePrefixMatch`,
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
		ErrorCases("invalid backendRefs",
			[]validators.Violation{{
				Field:   `spec.to[0].rules[0].default.backendRefs[0].name`,
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
          type: PathPrefix
          value: /
      default:
        backendRefs:
          - kind: MeshServiceSubset
            tags:
              version: v1

`),
		ErrorCases("hostnames and hostname to backend rewrite not allowed with services",
			[]validators.Violation{{
				Field:   `spec.to[0].hostnames`,
				Message: `must not be defined`,
			}, {
				Field:   "spec.to[0].rules[0].default.filters[0].urlRewrite.hostToBackendHostname",
				Message: "can only be set with MeshGateway",
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
    name: backend
  hostnames:
    - backend.com
  rules:
    - matches:
      - path:
          type: PathPrefix
          value: /
      default:
        filters:
          - type: URLRewrite
            urlRewrite:
              hostToBackendHostname: true
        backendRefs:
          - kind: MeshService
            name: backend
`),
		ErrorCase("top level MeshGateway requires backendRefs",
			validators.Violation{
				Field:   `spec.to[0].rules[0].default.backendRefs`,
				Message: `must not be empty`,
			}, `
type: MeshHTTPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshGateway
  name: edge
to:
- targetRef:
    kind: Mesh
  rules:
    - matches:
      - path:
          type: PathPrefix
          value: /
      default:
        filters:
          - type: URLRewrite
            urlRewrite:
              path:
                type: ReplacePrefixMatch
                replacePrefixMatch: /other
`),
		ErrorCases("invalid backendRef in requestMirror",
			[]validators.Violation{{
				Field:   `spec.to[0].rules[0].default.filters[0].requestMirror.backendRef.name`,
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
          type: PathPrefix
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
		ErrorCases("invalid hostnames",
			[]validators.Violation{
				{
					Field:   "spec.to[0].rules[0].default.filters[0].requestRedirect.hostname",
					Message: "cannot be an IP address",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[1].requestRedirect.hostname",
					Message: "cannot be an IP address",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[2].requestRedirect.hostname",
					Message: "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[3].requestRedirect.hostname",
					Message: "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[4].requestRedirect.hostname",
					Message: "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[5].requestRedirect.hostname",
					Message: "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[6].requestRedirect.hostname",
					Message: "must be no more than 253 characters",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[7].requestRedirect.hostname",
					Message: "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[8].requestRedirect.hostname",
					Message: "must be no more than 253 characters",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[8].requestRedirect.hostname",
					Message: "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[9].urlRewrite.hostname",
					Message: "cannot be an IP address",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[10].urlRewrite.hostname",
					Message: "cannot be an IP address",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[11].urlRewrite.hostname",
					Message: "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[12].urlRewrite.hostname",
					Message: "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[13].urlRewrite.hostname",
					Message: "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[14].urlRewrite.hostname",
					Message: "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[15].urlRewrite.hostname",
					Message: "must be no more than 253 characters",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[16].urlRewrite.hostname",
					Message: "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[17].urlRewrite.hostname",
					Message: "must be no more than 253 characters",
				},
				{
					Field:   "spec.to[0].rules[0].default.filters[17].urlRewrite.hostname",
					Message: "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')",
				},
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
          type: PathPrefix
          value: /
      default:
        filters:
          - type: RequestRedirect
            requestRedirect:
              hostname: 127.0.0.1
          - type: RequestRedirect
            requestRedirect:
              hostname: 2001:db8:3333:4444:5555:6666:7777:8888
          - type: RequestRedirect
            requestRedirect:
              hostname: a..bc
          - type: RequestRedirect
            requestRedirect:
              hostname: ec2-35-160-210-253.us-west-2-.compute.amazonaws.com
          - type: RequestRedirect
            requestRedirect:
              hostname: -ec2_35$160%210-253.us-west-2-.compute.amazonaws.com
          - type: RequestRedirect
            requestRedirect:
              hostname: ec2-35-160-210-253.us-west-2-.compute.amazonaws.com.
          - type: RequestRedirect
            requestRedirect:
              hostname: a23456789-a23456789-a234567890.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.com
          - type: RequestRedirect
            requestRedirect:
              hostname: mx.foo.com.
          - type: RequestRedirect
            requestRedirect:
              hostname: a23456789-a23456789-a234567890.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a2345678.com.
          - type: URLRewrite
            urlRewrite:
              hostname: 127.0.0.1
          - type: URLRewrite
            urlRewrite:
              hostname: 2001:db8:3333:4444:5555:6666:7777:8888
          - type: URLRewrite
            urlRewrite:
              hostname: a..bc
          - type: URLRewrite
            urlRewrite:
              hostname: ec2-35-160-210-253.us-west-2-.compute.amazonaws.com
          - type: URLRewrite
            urlRewrite:
              hostname: -ec2_35$160%210-253.us-west-2-.compute.amazonaws.com
          - type: URLRewrite
            urlRewrite:
              hostname: ec2-35-160-210-253.us-west-2-.compute.amazonaws.com.
          - type: URLRewrite
            urlRewrite:
              hostname: a23456789-a23456789-a234567890.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.com
          - type: URLRewrite
            urlRewrite:
              hostname: mx.foo.com.
          - type: URLRewrite
            urlRewrite:
              hostname: a23456789-a23456789-a234567890.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a23456789.a2345678.com.
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
          type: PathPrefix
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
		Entry("MeshGateway to Mesh allowed", `
type: MeshHTTPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshGateway
  name: edge
to:
- targetRef:
    kind: Mesh
  rules:
    - matches:
      - path:
          value: /
          type: PathPrefix
      default:
        filters:
          - type: URLRewrite
            urlRewrite:
              hostToBackendHostname: true
        backendRefs:
          - kind: MeshService
            name: backend
`),
		Entry("MeshGateway with hostnames allowed", `
type: MeshHTTPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: MeshGateway
  name: edge
to:
- targetRef:
    kind: Mesh
  hostnames:
    - exammple.com
  rules:
    - matches:
      - path:
          value: /
          type: PathPrefix
      default:
        backendRefs:
          - kind: MeshService
            name: backend
`),
		Entry("MeshService and MeshMultiZoneService", `
type: MeshHTTPRoute
mesh: mesh-1
name: route-1
targetRef:
  kind: Mesh
to:
- targetRef:
    kind: MeshService
    labels:
      kuma.io/display-name: backend
  rules:
    - matches:
      - path:
          value: /v1
          type: PathPrefix
      default:
        backendRefs:
          - kind: MeshService
            labels:
              kuma.io/display-name: backend
            port: 8080
    - matches:
      - path:
          value: /v2
          type: PathPrefix
      default:
        backendRefs:
          - kind: MeshMultiZoneService
            labels:
              kuma.io/display-name: backend
            port: 8080
`),
	)
})
