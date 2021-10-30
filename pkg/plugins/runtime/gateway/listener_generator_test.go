package gateway_test

import (
	"path"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/test/matchers"
	xds_server "github.com/kumahq/kuma/pkg/xds/server/v3"
)

var _ = Describe("Gateway Listener", func() {
	var rt runtime.Runtime

	Do := func(gateway string) (cache.Snapshot, error) {
		serverCtx := xds_server.NewXdsContext()
		reconciler := xds_server.DefaultReconciler(rt, serverCtx)

		Expect(StoreInlineFixture(rt, []byte(gateway))).To(Succeed())

		// Unmarshal the gateway YAML again so that we can figure
		// out which mesh it's in.
		r, err := rest.UnmarshallToCore([]byte(gateway))
		Expect(err).To(Succeed())

		// We expect there to be a Dataplane fixture named
		// "default" in the current mesh.
		ctx, proxy := MakeGeneratorContext(rt,
			core_model.ResourceKey{Mesh: r.GetMeta().GetMesh(), Name: "default"})

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

		Expect(StoreNamedFixture(rt, "mesh-tracing.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "dataplane-tracing.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "traffictrace.yaml")).To(Succeed())

		Expect(StoreNamedFixture(rt, "mesh-logging.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "dataplane-logging.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "trafficlog.yaml")).To(Succeed())
	})

	DescribeTable("Generate Envoy xDS resources",
		func(golden string, gateway string) {
			snap, err := Do(gateway)
			Expect(err).To(Succeed())

			out, err := yaml.Marshal(MakeProtoResource(snap.Resources[envoy_types.Listener]))
			Expect(err).To(Succeed())

			Expect(out).To(matchers.MatchGoldenYAML(path.Join("testdata", golden)))
		},
		Entry("should generate a single listener",
			"01-gateway-listener.yaml", `
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
`),
		Entry("should generate a multiple listeners",
			"02-gateway-listener.yaml", `
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
  - port: 9090
    protocol: HTTP
    tags:
      port: http/9090
`),
		Entry("should generate listener tracing",
			"03-gateway-listener.yaml", `
type: Gateway
mesh: tracing
name: tracing-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    tags:
      port: http/8080
`),
		Entry("should generate listener logging",
			"04-gateway-listener.yaml", `
type: Gateway
mesh: logging
name: logging-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    tags:
      port: http/8080
`),
	)

	DescribeTable("fail to generate xDS resources",
		func(errMsg string, gateway string) {
			_, err := Do(gateway)
			Expect(err).ToNot(Succeed())
			Expect(err.Error()).To(ContainSubstring(errMsg))
		},

		Entry("incompatible listeners",
			"cannot collapse listener protocols", `
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
  - port: 8080
    protocol: HTTPS
    tags:
      port: http/9090
`,
		),

		Entry("unsupported protocol",
			"unsupported protocol", `
type: Gateway
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 443
    protocol: HTTPS
    tags:
      port: http/443
`,
		),
	)
})
