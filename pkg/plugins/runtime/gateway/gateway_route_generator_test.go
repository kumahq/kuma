package gateway_test

import (
	"context"
	"path"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/xds"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	xds_server "github.com/kumahq/kuma/pkg/xds/server/v3"
)

type WithoutResource struct {
	Resource core_model.ResourceType
	Mesh     string
	Name     string
}

var _ = Describe("Gateway Route", func() {
	var rt runtime.Runtime
	var dataplanes *DataplaneGenerator

	Do := func() (cache.ResourceSnapshot, error) {
		serverCtx := xds_server.NewXdsContext()
		statsCallbacks, err := util_xds.NewStatsCallbacks(rt.Metrics(), "xds")
		if err != nil {
			return nil, err
		}
		reconciler := xds_server.DefaultReconciler(rt, serverCtx, statsCallbacks)

		// We expect there to be a Dataplane fixture named
		// "default" in the current mesh.
		ctx, proxy := MakeGeneratorContext(rt,
			core_model.ResourceKey{Mesh: "default", Name: "default"})

		// Tokens for zone egress needs to be generated
		// Without test configuration each run will generates
		// new tokens for authentication.
		ctx.ControlPlane.Secrets = &xds.TestSecrets{}

		Expect(proxy.Dataplane.Spec.IsBuiltinGateway()).To(BeTrue())

		if err := reconciler.Reconcile(*ctx, proxy); err != nil {
			return nil, err
		}

		return serverCtx.Cache().GetSnapshot(proxy.Id.String())
	}

	BeforeEach(func() {
		var err error

		rt, err = BuildRuntime()
		Expect(err).To(Succeed(), "build runtime instance")

		Expect(StoreNamedFixture(rt, "mesh-default.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "serviceinsight-default.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "dataplane-default.yaml")).To(Succeed())

		dataplanes = &DataplaneGenerator{Manager: rt.ResourceManager()}

		// Add dataplane resources for all the services used in the test suite.
		for _, service := range []string{
			"api-service",
			"echo-exact",
			"echo-mirror",
			"echo-prefix",
			"echo-regex",
			"echo-service",
			"exact-header-match",
			"exact-query-match",
			"exact-service",
			"prefix-service",
			"regex-header-match",
			"regex-query-match",
		} {
			dataplanes.Generate(service)
		}
	})

	entries := []TableEntry{
		// When we have a route with multiple hostnames that is
		// attached to a listener with no hostname (i.e. a wildcard),
		// we ought to generate distinct Envoy virtualhost entries
		// for each hostname
		Entry("should expand route hostnames into virtual hosts",
			"01-gateway-route.yaml", `
type: MeshGateway
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    tags:
      port: http/8080
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    hostnames:
    - foo.example.com
    - bar.example.com
    - "*.example.com"
    rules:
    - matches:
      - path:
          match: EXACT
          value: /
      backends:
      - destination:
          kuma.io/service: echo-service
`,
		),

		// Attaching a wildcard route (i.e. configured without
		// any hostnames) to a wildcard Gateway listener should
		// generate a wildcard virtual host.
		Entry("should create a wildcard virtual host",
			"02-gateway-route.yaml", `
type: MeshGateway
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    tags:
      port: http/8080
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: EXACT
          value: /
      backends:
      - destination:
          kuma.io/service: echo-service
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service-extra
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    hostnames:
    - extra.example.com
    rules:
    - matches:
      - path:
          match: EXACT
          value: /
      backends:
      - destination:
          kuma.io/service: echo-service
`,
		),

		// Given a route with matching hostnames and a route with
		// non-matching hostnames, only the route with matching
		// hostnames should be configured on the Gateway.
		Entry("should match route hostnames on the listener",
			"03-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    hostnames:
    - echo.example.com
    rules:
    - matches:
      - path:
          match: EXACT
          value: /service/echo
      backends:
      - destination:
          kuma.io/service: echo-service
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service-2
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    hostnames:
    - foo.example.com
    rules:
    - matches:
      - path:
          match: EXACT
          value: /service/foo
      backends:
      - destination:
          kuma.io/service: echo-service
`,
		),

		Entry("should create a mirror route",
			"04-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: EXACT
          value: /
      filters:
      - mirror:
          percentage: 0.001
          backend:
            destination:
              kuma.io/service: echo-mirror
              kuma.io/zone: development
      backends:
      - destination:
          kuma.io/service: echo-service
`,
		),

		Entry("should create a redirect route",
			"05-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: EXACT
          value: /
      filters:
      - redirect:
          scheme: https
          hostname: example.com
          status_code: 302
`,
		),

		Entry("should create a request header rewrite route",
			"06-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: EXACT
          value: /
      filters:
      - mirror:
          percentage: 0.001
          backend:
            destination:
              kuma.io/service: echo-mirror
              kuma.io/zone: development
      - request_header:
          set:
          - name: set-header-one
            value : one
          - name: set-header-two
            value: two
          add:
          - name: append-forwarded
            value: "yes"
          - name: append-forwarded
            value: please
          remove:
          - delete-another
      backends:
      - destination:
          kuma.io/service: echo-service
`,
		),

		Entry("should create a path prefix match",
			"07-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: /api
      backends:
      - destination:
          kuma.io/service: echo-service
`,
		),

		Entry("should disambiguate path prefix and exact matches",
			"08-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: /api
      backends:
      - destination:
          kuma.io/service: prefix-service
    - matches:
      - path:
          match: EXACT
          value: /api
      backends:
      - destination:
          kuma.io/service: exact-service
`,
		),

		Entry("should create a regex prefix match",
			"09-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: REGEX
          value: "^/api/v[0-9]+$"
      backends:
      - destination:
          kuma.io/service: echo-service
`,
		),

		Entry("should create a header route match",
			"10-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - headers:
        - match: EXACT
          name: Content-Type
          value: 'application/json'
      - headers:
        - match: EXACT
          name: Language
          value: gibberish
      backends:
      - destination:
          kuma.io/service: exact-header-match
    - matches:
      - headers:
        - match: REGEX
          name: Content-Type
          value: 'application/.*'
        - match: REGEX
          name: Language
          value: '.*sh'
      backends:
      - destination:
          kuma.io/service: regex-header-match
`,
		),

		Entry("should create a query param route match",
			"11-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - query_parameters:
        - match: EXACT
          name: Content-Type
          value: 'application/json'
      - query_parameters:
        - match: EXACT
          name: Language
          value: gibberish
      backends:
      - destination:
          kuma.io/service: exact-query-match
    - matches:
      - query_parameters:
        - match: REGEX
          name: Content-Type
          value: 'application/.*'
        - match: REGEX
          name: Language
          value: '.*sh'
      backends:
      - destination:
          kuma.io/service: regex-query-match
`,
		),

		Entry("duplicates routes for repeated matches",
			"12-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - headers:
        - match: EXACT
          name: Content-Type
          value: 'application/json'
        path:
          match: EXACT
          value: /app/json
        method: PUT
      - headers:
        - match: EXACT
          name: Language
          value: gibberish
        path:
          match: PREFIX
          value: /lang/json
        method: GET
      backends:
      - destination:
          kuma.io/service: echo-service
`,
		),

		Entry("order path matches by specificity",
			"13-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: REGEX
          value: /match/foo
      backends:
      - destination:
          kuma.io/service: echo-regex
    - matches:
      - path:
          match: EXACT
          value: /match/bar
      backends:
      - destination:
          kuma.io/service: echo-exact
    - matches:
      - path:
          match: PREFIX
          value: /match/baz
      backends:
      - destination:
          kuma.io/service: echo-prefix
`,
		),

		Entry("rewrite the host header",
			"14-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: EXACT
          value: /app/json
        method: PUT
      filters:
      - request_header:
          set:
          - name: Host
            value: newhost.example.com
      backends:
      - destination:
          kuma.io/service: echo-service
`,
		),

		Entry("match timeout policy",
			"15-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: /
      filters:
      - mirror:
          percentage: 1
          backend:
            destination:
              kuma.io/service: echo-mirror
      backends:
      - destination:
          kuma.io/service: echo-service
    - matches:
      - path:
          match: PREFIX
          value: /api
      backends:
      - destination:
          kuma.io/service: api-service
`, `
type: Timeout
mesh: default
name: echo-service
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: echo-service
conf:
  connect_timeout: 10s
  http:
    request_timeout: 10s
    idle_timeout: 10s
`, `
type: Timeout
mesh: default
name: api-service
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: api-service
conf:
  connect_timeout: 20s
  http:
    request_timeout: 20s
    idle_timeout: 20s
`, `
type: Timeout
mesh: default
name: echo-mirror
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: echo-mirror
conf:
  connect_timeout: 300s
  http:
    request_timeout: 30s
    idle_timeout: 30s
`,
		),

		Entry("match circuit breaker policy",
			"16-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: /
      filters:
      - mirror:
          percentage: 1
          backend:
            destination:
              kuma.io/service: echo-mirror
      backends:
      - destination:
          kuma.io/service: echo-service
    - matches:
      - path:
          match: PREFIX
          value: /api
      backends:
      - destination:
          kuma.io/service: api-service
`, `
type: CircuitBreaker
mesh: default
name: echo-service
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: echo-service
conf:
  baseEjectionTime: 10s
  thresholds:
    maxRetries: 10
  detectors:
    localErrors:
      consecutive: 10
`, `
type: CircuitBreaker
mesh: default
name: api-service
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: api-service
conf:
  baseEjectionTime: 20s
  thresholds:
    maxRetries: 20
  detectors:
    localErrors:
      consecutive: 20
`, `
type: CircuitBreaker
mesh: default
name: echo-mirror
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: echo-mirror
conf:
  baseEjectionTime: 30s
  thresholds:
    maxRetries: 20
  detectors:
    localErrors:
      consecutive: 30
`,
		),

		Entry("match health check policy",
			"17-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: /
      filters:
      - mirror:
          percentage: 1
          backend:
            destination:
              kuma.io/service: echo-mirror
      backends:
      - destination:
          kuma.io/service: echo-service
    - matches:
      - path:
          match: PREFIX
          value: /api
      backends:
      - destination:
          kuma.io/service: api-service
`, `
type: HealthCheck
mesh: default
name: echo-service
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: echo-service
conf:
  interval: 10s
  timeout: 10s
  healthyThreshold: 1
  unhealthyThreshold: 1
`, `
type: HealthCheck
mesh: default
name: api-service
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: api-service
conf:
  interval: 20s
  timeout: 20s
  healthyThreshold: 2
  unhealthyThreshold: 2
`, `
type: HealthCheck
mesh: default
name: echo-mirror
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: echo-mirror
conf:
  interval: 30s
  timeout: 30s
  healthyThreshold: 3
  unhealthyThreshold: 3
`,
		),

		Entry("generates a HTTP external service cluster",
			"18-gateway-route.yaml", `
type: ExternalService
mesh: default
name: external-httpbin
tags:
  kuma.io/service: external-httpbin
  kuma.io/protocol: http
networking:
  address: httpbin.com:80
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: "/"
      backends:
      - destination:
          kuma.io/service: external-httpbin
`,
		),

		Entry("generates a HTTP/2 external service cluster",
			"19-gateway-route.yaml", `
type: ExternalService
mesh: default
name: external-httpbin
tags:
  kuma.io/service: external-httpbin
  kuma.io/protocol: http2
networking:
  address: httpbin.com:443
  tls:
    enabled: true
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: "/"
      backends:
      - destination:
          kuma.io/service: external-httpbin
`,
		),

		Entry("generates policy-specific clusters",
			"20-gateway-route.yaml",
			`
# Rewrite the dataplane to attach the "gateway-multihost" Gateway.
type: Dataplane
mesh: default
name: default
networking:
  address: 192.168.1.1
  gateway:
    type: BUILTIN
    tags:
      kuma.io/service: gateway-multihost
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-multihost
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: /
      backends:
      - destination:
          kuma.io/service: echo-service
`, `
type: Timeout
mesh: default
name: host-one-timeout
sources:
- match:
    kuma.io/service: gateway-multihost
    hostname: one.example.com
destinations:
- match:
    kuma.io/service: echo-service
conf:
  connect_timeout: 10s
  http:
    request_timeout: 10s
    idle_timeout: 10s
`, `
type: Timeout
mesh: default
name: host-two-timeout
sources:
- match:
    kuma.io/service: gateway-multihost
    hostname: two.example.com
destinations:
- match:
    kuma.io/service: echo-service
conf:
  connect_timeout: 20s
  http:
    request_timeout: 20s
    idle_timeout: 20s
`, `
type: Timeout
mesh: default
name: host-three-timeout
sources:
- match:
    kuma.io/service: gateway-multihost
    hostname: three.example.com
destinations:
- match:
    kuma.io/service: echo-service
conf:
  connect_timeout: 300s
  http:
    request_timeout: 30s
    idle_timeout: 30s
`,
		),

		Entry("match ratelimit policy",
			"21-gateway-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: /
      filters:
      - mirror:
          percentage: 1
          backend:
            destination:
              kuma.io/service: echo-mirror
      backends:
      - destination:
          kuma.io/service: echo-service
    - matches:
      - path:
          match: PREFIX
          value: /api
      backends:
      - destination:
          kuma.io/service: api-service
`, `
type: RateLimit
mesh: default
name: echo-service
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: echo-service
conf:
  http:
    requests: 1
    interval: 10s
`, `
type: RateLimit
mesh: default
name: api-service
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: api-service
conf:
  http:
    requests: 1
    interval: 20s
`, `
# This does nothing because rate limits are per-route, not per-cluster.
type: RateLimit
mesh: default
name: echo-mirror
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: echo-mirror
conf:
  http:
    requests: 1
    interval: 30s
`,
		),

		Entry("should distribute routes across wildcard listener",
			"22-gateway-route.yaml", `
type: MeshGateway
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    hostname: example.com
    tags:
      hostname: example.com
      port: http/8080
  - port: 8080
    protocol: HTTP
    tags:
      hostname: all
      port: http/8080
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
    hostname: example.com
conf:
  http:
    rules:
    - matches:
      - path:
          match: EXACT
          value: /v2
      backends:
      - destination:
          kuma.io/service: api-service
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service-extra
selectors:
- match:
    kuma.io/service: gateway-default
    hostname: all
conf:
  http:
    hostnames:
    - "example.com"
    rules:
    - matches:
      - path:
          match: EXACT
          value: /
      backends:
      - destination:
          kuma.io/service: echo-service
`,
		),

		Entry("should distribute partial matching routes across wildcard listener",
			"23-gateway-route.yaml", `
type: MeshGateway
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    hostname: example.com
    tags:
      hostname: example.com
      port: http/8080
  - port: 8080
    protocol: HTTP
    tags:
      hostname: all
      port: http/8080
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
    hostname: example.com
conf:
  http:
    rules:
    - matches:
      - path:
          match: EXACT
          value: /v2
      backends:
      - destination:
          kuma.io/service: api-service
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service-extra
selectors:
- match:
    kuma.io/service: gateway-default
    hostname: all
conf:
  http:
    hostnames:
    - "*.com"
    rules:
    - matches:
      - path:
          match: EXACT
          value: /
      backends:
      - destination:
          kuma.io/service: echo-service
`,
		),

		Entry("should distribute partial matching routes for one listener",
			"24-gateway-route.yaml", `
type: MeshGateway
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    tags:
      port: http/8080
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    hostnames:
    - "example.com"
    rules:
    - matches:
      - path:
          match: EXACT
          value: /v2
      backends:
      - destination:
          kuma.io/service: api-service
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service-extra
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    hostnames:
    - "*.com"
    rules:
    - matches:
      - path:
          match: EXACT
          value: /
      backends:
      - destination:
          kuma.io/service: echo-service
`,
		),

		Entry("generates external service endpoints with zone egress ip when zone egress enabled and zone egress instances available",
			"25-gateway-route.yaml", `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
routing:
  zoneEgress: true`, `
type: ExternalService
mesh: default
name: external-httpbin
tags:
  kuma.io/service: external-httpbin
  kuma.io/protocol: http2
networking:
  address: httpbin.com:443
  tls:
    enabled: true
`, `
type: ZoneEgress
name: zone-egress
networking:
  address: 1.1.1.1
  port: 12345
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: "/"
      backends:
      - destination:
          kuma.io/service: external-httpbin
`,
		),

		Entry("generates external service cluster without endpoint ip because there is no zone egress instance",
			"26-gateway-route.yaml", `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
routing:
  zoneEgress: true`, `
type: ExternalService
mesh: default
name: external-httpbin
tags:
  kuma.io/service: external-httpbin
  kuma.io/protocol: http2
networking:
  address: httpbin.com:443
  tls:
    enabled: true
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: "/"
      backends:
      - destination:
          kuma.io/service: external-httpbin
`,
		),

		Entry("generates cross mesh gateway listeners",
			"cross-mesh-gateway.yaml", `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`, `
type: Mesh
name: other
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`, `
type: ExternalService
mesh: default
name: external-httpbin
tags:
  kuma.io/service: external-httpbin
  kuma.io/protocol: http2
networking:
  address: httpbin.com:443
  tls:
    enabled: true
`, `
type: MeshGateway
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    crossMesh: true
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: "/ext"
      backends:
      - destination:
          kuma.io/service: external-httpbin
    - matches:
      - path:
          match: PREFIX
          value: "/echo"
      backends:
      - destination:
          kuma.io/service: echo-service
`,
		),

		Entry("works with no Timeout policy",
			"no-timeout.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: EXACT
          value: /
      backends:
      - destination:
          kuma.io/service: echo-service
`,
			WithoutResource{
				Resource: core_mesh.TimeoutType,
				Mesh:     "default",
				Name:     "timeout-all-default",
			},
		),

		Entry("timeout policy works with external services without egress",
			"external-service-with-timeout-no-egress.yaml", `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`, `
type: ExternalService
mesh: default
name: external-httpbin
tags:
  kuma.io/service: external-httpbin
  kuma.io/protocol: http2
networking:
  address: httpbin.com:443
  tls:
    enabled: true
`, `
type: MeshGateway
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: HTTP
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: "/ext"
      backends:
      - destination:
          kuma.io/service: external-httpbin
`, `
type: Timeout
mesh: default
name: es-timeouts
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: external-httpbin
conf:
  connect_timeout: 113s
  http:
    request_timeout: 114s
    idle_timeout: 115s
    stream_idle_timeout: 116s
    max_stream_duration: 117s
`,
		),

		Entry("doesn't create invalid config with tcp route",
			"http-tcp-route.yaml", `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  tcp:
    rules:
    - backends:
      - destination:
          kuma.io/service: service
`,
		),
	}

	tcpEntries := []TableEntry{
		Entry("generates clusters for TCP",
			"tcp-route.yaml", `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
routing:
  zoneEgress: true
`, `
type: MeshGateway
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: TCP
    tags:
      port: http/8080
`, `
type: ExternalService
mesh: default
name: external-httpbin
tags:
  kuma.io/service: external-httpbin
networking:
  address: httpbin.com:443
  tls:
    enabled: true
`, `
type: MeshGatewayRoute
mesh: default
name: external-or-api
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  tcp:
    rules:
    - backends:
      - destination:
          kuma.io/service: external-httpbin
      - destination:
          kuma.io/service: api-service
`, `
type: MeshGatewayRoute
mesh: default
name: echo-service
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  tcp:
    rules:
    - backends:
      - destination:
          kuma.io/service: echo-service
`, `
type: Timeout
mesh: default
name: echo-service
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: '*'
conf:
  connect_timeout: 10s
  http:
    request_timeout: 10s
    idle_timeout: 10s
`,
		),
		Entry("generates direct cluster for TCP external service",
			"tcp-route-no-egress.yaml", `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
routing:
  zoneEgress: false
`, `
type: MeshGateway
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: TCP
    tags:
      port: http/8080
`, `
type: ExternalService
mesh: default
name: external-httpbin
tags:
  kuma.io/service: external-httpbin
networking:
  address: httpbin.com:443
  tls:
    enabled: true
`, `
type: MeshGatewayRoute
mesh: default
name: external-or-api
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  tcp:
    rules:
    - backends:
      - destination:
          kuma.io/service: external-httpbin
`, `
type: Timeout
mesh: default
name: echo-service
sources:
- match:
    kuma.io/service: gateway-default
destinations:
- match:
    kuma.io/service: '*'
conf:
  connect_timeout: 12s
  http:
    request_timeout: 10s
    idle_timeout: 10s
`,
		),
	}

	handleArg := func(arg interface{}) {
		switch val := arg.(type) {
		case WithoutResource:
			obj, err := registry.Global().NewObject(val.Resource)
			Expect(err).ToNot(HaveOccurred())
			Expect(rt.ResourceManager().Delete(
				context.Background(), obj, store.DeleteByKey(val.Name, val.Mesh),
			)).To(Succeed())
		case string:
			Expect(StoreInlineFixture(rt, []byte(val))).To(Succeed())
		}
	}
	Context("with a HTTP gateway", func() {
		JustBeforeEach(func() {
			Expect(StoreNamedFixture(rt, "gateway-http-multihost.yaml")).To(Succeed())
			Expect(StoreNamedFixture(rt, "gateway-http-default.yaml")).To(Succeed())
		})
		DescribeTable("generating xDS resources",
			func(goldenFileName string, args ...interface{}) {
				// given
				for _, arg := range args {
					handleArg(arg)
				}

				// when
				snap, err := Do()
				Expect(err).To(Succeed())

				// then
				Expect(yaml.Marshal(MakeProtoSnapshot(snap))).
					To(matchers.MatchGoldenYAML(path.Join("testdata", "http", goldenFileName)))

				// then
				Expect(snap.(*cache.Snapshot).Consistent()).To(Succeed())
			},
			entries,
		)
	})

	Context("with a HTTPS gateway", func() {
		JustBeforeEach(func() {
			Expect(StoreNamedFixture(rt, "gateway-https-multihost.yaml")).To(Succeed())
			Expect(StoreNamedFixture(rt, "gateway-https-default.yaml")).To(Succeed())
			Expect(StoreNamedFixture(rt, "secret-https-default.yaml")).To(Succeed())
		})
		DescribeTable("generating xDS resources",
			func(goldenFileName string, args ...interface{}) {
				// given
				for _, arg := range args {
					handleArg(arg)
				}

				// when
				snap, err := Do()
				Expect(err).To(Succeed())

				// then
				Expect(yaml.Marshal(MakeProtoSnapshot(snap))).
					To(matchers.MatchGoldenYAML(path.Join("testdata", "https", goldenFileName)))

			},
			entries,
		)
	})

	Context("with a TCP gateway", func() {
		DescribeTable("generating xDS resources",
			func(goldenFileName string, fixtureResources ...string) {
				// given
				for _, resource := range fixtureResources {
					Expect(StoreInlineFixture(rt, []byte(resource))).To(Succeed())
				}

				// when
				snap, err := Do()
				Expect(err).To(Succeed())

				// then
				Expect(yaml.Marshal(MakeProtoSnapshot(snap))).
					To(matchers.MatchGoldenYAML(path.Join("testdata", "tcp", goldenFileName)))

			},
			tcpEntries,
		)
	})
})
