package mesh_test

import (
	. "github.com/onsi/ginkgo/v2"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
	_ "github.com/kumahq/kuma/pkg/plugins/runtime/gateway/register"
)

// MeshGatewayRouteGenerator is a ResourceGenerator that creates MeshGatewayResource objects.
type MeshGatewayRouteGenerator func() *MeshGatewayRouteResource

func (g MeshGatewayRouteGenerator) New() model.Resource {
	if g != nil {
		return g()
	}

	return nil
}

var _ = Describe("MeshGatewayRoute", func() {
	DescribeValidCases(MeshGatewayRouteGenerator(NewMeshGatewayRouteResource),

		Entry("HTTP route", `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    hostnames:
    - foo.example.com
    rules:
    - matches:
      - method: GET
        path:
          match: EXACT
          value: /
        headers:
        - match: EXACT
          name: x-foo
          value: "my great foo"
        - match: REGEX
          name: x-count
          value: "[0-9]+"
        query_parameters:
        - match: EXACT
          name: customer
          value: kong
      filters:
      - request_header:
          set:
          - name: x-foo
            value: foo
          - name: x-foo
            value: foo
          add:
          - name: x-added
            value: added
          - name: x-added
            value: added
          remove:
          - x-deleted
      - mirror:
         percentage: 1.12
         backend:
           destination:
             kuma.io/service: target
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-1
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		Entry("HTTP redirect", `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    hostnames:
    - foo.example.com
    rules:
    - matches:
      - method: GET
        path:
          match: EXACT
          value: /
        headers:
        - match: EXACT
          name: x-foo
          value: "my great foo"
        - match: REGEX
          name: x-count
          value: "[0-9]+"
        query_parameters:
        - match: EXACT
          name: customer
          value: kong
      filters:
      - redirect:
         scheme: http
         hostname: foo.example.com
         port: 80
         status_code: 307
`),
	)

	DescribeErrorCases(MeshGatewayRouteGenerator(NewMeshGatewayRouteResource),
		ErrorCase("missing conf", validators.Violation{
			Field:   "conf",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
`),
		ErrorCase("missing HTTP rules", validators.Violation{
			Field:   "conf.http.rules",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules: []
`),
		ErrorCase("missing HTTP rule matches", validators.Violation{
			Field:   "conf.http.rules[0].matches",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("missing HTTP rule backends", validators.Violation{
			Field:   "conf.http.rules[0].backends",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          match: EXACT
          value: /
`),
		ErrorCase("backends with no service", validators.Violation{
			Field:   "conf.http.rules[0].backends[0]",
			Message: `mandatory tag "kuma.io/service" is missing`,
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          match: EXACT
          value: /
      backends:
      - weight: 5
        destination:
          phoney: target-2
`),
		ErrorCase("rule with missing match", validators.Violation{
			Field:   "conf.http.rules[0].matches[0]",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - {}
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("match with empty path", validators.Violation{
			Field:   "conf.http.rules[0].matches[0].value",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          match: REGEX
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("match with relative path", validators.Violation{
			Field:   "conf.http.rules[0].matches[0].value",
			Message: "must be an absolute path",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          value: some/path
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("match with empty header name", validators.Violation{
			Field:   "conf.http.rules[0].matches[0].headers[0].name",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - headers:
        - value: value
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("match with empty header value", validators.Violation{
			Field:   "conf.http.rules[0].matches[0].headers[0].value",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - headers:
        - name: value
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("match with empty query name", validators.Violation{
			Field:   "conf.http.rules[0].matches[0].query_parameters[0].name",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - query_parameters:
        - value: value
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("match with empty query value", validators.Violation{
			Field:   "conf.http.rules[0].matches[0].query_parameters[0].value",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - query_parameters:
        - name: value
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("empty request header filter", validators.Violation{
			Field:   "conf.http.rules[0].filters[0].request_header",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          value: /
      filters:
      - request_header: {}
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("empty request header filter set", validators.Violation{
			Field:   "conf.http.rules[0].filters[0].request_header.set[0].value",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          value: /
      filters:
      - request_header:
          set:
          - name: x-foo
          add:
          - name: x-foo
            value: foo
          remove:
          - foo
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("empty request header filter add", validators.Violation{
			Field:   "conf.http.rules[0].filters[0].request_header.add[0].name",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          value: /
      filters:
      - request_header:
          set:
          - name: x-foo
            value: foo
          add:
          - value: foo
          remove:
          - foo
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("empty request header filter remove", validators.Violation{
			Field:   "conf.http.rules[0].filters[0].request_header.remove[0]",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          value: /
      filters:
      - request_header:
          set:
          - name: x-foo
            value: foo
          add:
          - name: x-foo
            value: foo
          remove:
          - ""
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("empty mirror filter backend", validators.Violation{
			Field:   "conf.http.rules[0].filters[0].mirror.backend",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          value: /
      filters:
      - mirror:
          percentage: 0.0
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("mirror filter invalid percentage", validators.Violation{
			Field:   "conf.http.rules[0].filters[0].mirror.percentage",
			Message: "has to be in [0.0 - 100.0] range",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          value: /
      filters:
      - mirror:
          percentage: -1.0
          backend:
            destination:
              kuma.io/service: target-2
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("redirect filter with empty scheme", validators.Violation{
			Field:   "conf.http.rules[0].filters[0].redirect.scheme",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          value: /
      filters:
      - redirect:
          hostname: example.com
          status_code: 301
`),
		ErrorCase("redirect filter with empty hostname", validators.Violation{
			Field:   "conf.http.rules[0].filters[0].redirect.hostname",
			Message: "cannot be empty",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          value: /
      filters:
      - redirect:
          scheme: https
          status_code: 301
`),
		ErrorCase("redirect filter with invalid port", validators.Violation{
			Field:   "conf.http.rules[0].filters[0].redirect.port",
			Message: "port must be in the range [1, 65535]",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          value: /
      filters:
      - redirect:
          scheme: https
          hostname: example.com
          port: 128555
          status_code: 301
`),
		ErrorCase("redirect filter with invalid status", validators.Violation{
			Field:   "conf.http.rules[0].filters[0].redirect.status_code",
			Message: "must be in the range [300, 308]",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          value: /
      filters:
      - redirect:
          scheme: https
          hostname: example.com
          status_code: 500
`),
		ErrorCase("redirect filter with backend routes", validators.Violation{
			Field:   "conf.http.rules[0].backends",
			Message: "must be empty when using redirect filters",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          value: /
      filters:
      - redirect:
          scheme: https
          hostname: example.com
          status_code: 300
      backends:
      - weight: 5
        destination:
          kuma.io/service: target-2
`),
		ErrorCase("redirect prevents other filters", validators.Violation{
			Field:   "conf.http.rules[0].filters",
			Message: "redirects cannot be used with other filters",
		}, `
type: MeshGatewayRoute
name: route
mesh: default
selectors:
- match:
    kuma.io/service: gateway
conf:
  http:
    rules:
    - matches:
      - path:
          value: /
      filters:
      - redirect:
          scheme: https
          hostname: example.com
          status_code: 300
      - mirror:
        backends:
        - weight: 5
          destination:
            kuma.io/service: target-2
`),
	)
})
