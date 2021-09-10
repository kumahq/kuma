package gateway_test

import (
	"path"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
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
	})

	It("should expand route hostnames into virtual hosts", func() {
		dataplanes.Generate("echo-service") // TODO(jpeach) not used yet

		// Given the default gateway has a wildcard listener.
		Expect(StoreInlineFixture(rt, []byte(`
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
`))).To(Succeed())

		// When we have a route with multiple hostnames that is
		// attached to a listener with no hostname (i.e. a wildcard),
		// we ought to generate distinct Envoy virtualhost entries
		// for each hostname
		Expect(StoreInlineFixture(rt, []byte(`
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
`))).To(Succeed())

		// when
		snap, err := Do()
		Expect(err).To(Succeed())

		// then
		Expect(yaml.Marshal(MakeProtoSnapshot(snap))).
			To(matchers.MatchGoldenYAML(path.Join("testdata", "01-gateway-route.yaml")))
	})

	It("should create a wildcard virtual host", func() {
		dataplanes.Generate("echo-service") // TODO(jpeach) not used yet

		// Given the default gateway has a wildcard listener.
		Expect(StoreInlineFixture(rt, []byte(`
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
`))).To(Succeed())

		// Given a route with no hostnames.
		Expect(StoreInlineFixture(rt, []byte(`
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
`))).To(Succeed())

		// Given a route with hostnames.
		Expect(StoreInlineFixture(rt, []byte(`
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
`))).To(Succeed())

		// when
		snap, err := Do()
		Expect(err).To(Succeed())

		// then
		Expect(yaml.Marshal(MakeProtoSnapshot(snap))).
			To(matchers.MatchGoldenYAML(path.Join("testdata", "02-gateway-route.yaml")))
	})

	It("should match route hostnames on the listener", func() {
		dataplanes.Generate("echo-service") // TODO(jpeach) not used yet

		// Given a route with matching hostnames.
		Expect(StoreInlineFixture(rt, []byte(`
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
`))).To(Succeed())

		// Given a route with non-matching hostnames.
		Expect(StoreInlineFixture(rt, []byte(`
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
`))).To(Succeed())

		// TODO(jpeach) This test won't test anything until we
		// implement route generation, because neither of the two
		// GatewayRoute fixtures generate anything at all

		// when
		snap, err := Do()
		Expect(err).To(Succeed())

		// then
		Expect(yaml.Marshal(MakeProtoSnapshot(snap))).
			To(matchers.MatchGoldenYAML(path.Join("testdata", "03-gateway-route.yaml")))
	})

})
