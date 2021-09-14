package gateway_test

import (
	"path"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/test/matchers"
	xds_server "github.com/kumahq/kuma/pkg/xds/server/v3"
)

var _ = Describe("Gateway Gateway Route", func() {
	var rt runtime.Runtime
	var dataplanes *DataplaneGenerator

	Do := func() (cache.Snapshot, error) {
		serverCtx := xds_server.NewXdsContext()
		reconciler := xds_server.DefaultReconciler(rt, serverCtx)

		// We expect there to be a Dataplane fixture named
		// "default" in the current mesh.
		ctx, proxy := MakeGeneratorContext(rt,
			core_model.ResourceKey{Mesh: "default", Name: "default"})

		Expect(proxy.Dataplane.Spec.IsBuiltinGateway()).To(BeTrue())

		if err := reconciler.Reconcile(*ctx, proxy); err != nil {
			return cache.Snapshot{}, err
		}

		return serverCtx.Cache().GetSnapshot(proxy.Id.String())
	}

	BeforeEach(func() {
		var err error

		rt, err = BuildRuntime()
		Expect(err).To(Succeed(), "build runtime instance")

		Expect(StoreNamedFixture(rt, "mesh-default.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "dataplane-default.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "gateway-default.yaml")).To(Succeed())

		dataplanes = &DataplaneGenerator{Manager: rt.ResourceManager()}

		// TODO(jpeach) dataplane resources won't be used in the
		// tests until the GatewayRoute generator starts generating
		// endpoints.
		dataplanes.Generate("echo-service")
	})

	DescribeTable("generate matching resources",
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
				To(matchers.MatchGoldenYAML(path.Join("testdata", goldenFileName)))

		},
		// When we have a route with multiple hostnames that is
		// attached to a listener with no hostname (i.e. a wildcard),
		// we ought to generate distinct Envoy virtualhost entries
		// for each hostname
		Entry("should expand route hostnames into virtual hosts",
			"01-gateway-route.yaml", `
type: Gateway
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
type: GatewayRoute
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
type: Gateway
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
type: GatewayRoute
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
type: GatewayRoute
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
		//
		// TODO(jpeach) This test won't test anything until we
		// implement route generation, because neither of the two
		// GatewayRoute fixtures generate anything at all
		Entry("should match route hostnames on the listener",
			"03-gateway-route.yaml", `
type: GatewayRoute
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
type: GatewayRoute
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
	)

})
