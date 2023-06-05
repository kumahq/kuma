package gateway_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Retry Generator", func() {
	// var rt runtime.Runtime
	// var dataplanes *DataplaneGenerator

	// Do := func() (cache.ResourceSnapshot, error) {
	// 	serverCtx := xds_server.NewXdsContext()
	// 	statsCallbacks, err := util_xds.NewStatsCallbacks(rt.Metrics(), "xds")
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	reconciler := xds_server.DefaultReconciler(rt, serverCtx, statsCallbacks)
	//
	// 	// We expect there to be a Dataplane fixture named[
	// 	// ]"default" in the current mesh.
	// 	ctx, proxy := MakeGeneratorContext(rt,
	// 		core_model.ResourceKey{Mesh: "default", Name: "default"})
	//
	// 	// Tokens for zone egress needs to be generated
	// 	// Without test configuration each run will generates
	// 	// new tokens for authentication.
	// 	ctx.ControlPlane.Secrets = &xds.TestSecrets{}
	//
	// 	Expect(proxy.Dataplane.Spec.IsBuiltinGateway()).To(BeTrue())
	//
	// 	if err := reconciler.Reconcile(*ctx, proxy); err != nil {
	// 		return nil, err
	// 	}
	//
	// 	return serverCtx.Cache().GetSnapshot(proxy.Id.String())
	// }

	// BeforeEach(func() {
	// 	var err error
	//
	// 	rt, err = BuildRuntime()
	// 	Expect(err).To(Succeed(), "build runtime instance")
	//
	// 	Expect(StoreNamedFixture(rt, "mesh-default.yaml")).To(Succeed())
	// 	Expect(StoreNamedFixture(rt, "serviceinsight-default.yaml")).To(Succeed())
	// 	Expect(StoreNamedFixture(rt, "dataplane-default.yaml")).To(Succeed())
	//
	// 	dataplanes = &DataplaneGenerator{Manager: rt.ResourceManager()}
	//
	// 	// Add dataplane resources for all the services used in the test suite.
	// 	for _, service := range []string{
	// 		"api-service",
	// 		"echo-exact",
	// 		"echo-mirror",
	// 		"echo-prefix",
	// 		"echo-regex",
	// 		"echo-service",
	// 		"exact-header-match",
	// 		"exact-query-match",
	// 		"exact-service",
	// 		"prefix-service",
	// 		"regex-header-match",
	// 		"regex-query-match",
	// 	} {
	// 		dataplanes.Generate(service)
	// 	}
	// })

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
	}

	// handleArg := func(arg interface{}) {
	// 	switch val := arg.(type) {
	// 	case WithoutResource:
	// 		obj, err := registry.Global().NewObject(val.Resource)
	// 		Expect(err).ToNot(HaveOccurred())
	// 		Expect(rt.ResourceManager().Delete(
	// 			context.Background(), obj, store.DeleteByKey(val.Name, val.Mesh),
	// 		)).To(Succeed())
	// 	case string:
	// 		Expect(StoreInlineFixture(rt, []byte(val))).To(Succeed())
	// 	}
	// }

	Context("with a HTTP gateway", func() {
		DescribeTable("generating xDS resources",
			func(goldenFileName string, args ...interface{}) {
				// given
				// for _, arg := range args {
				// 	handleArg(arg)
				// }

				// when
				// snap, err := Do()
				// Expect(err).To(Succeed())

				// then
				// Expect(yaml.Marshal(MakeProtoSnapshot(snap))).
				// 	To(matchers.MatchGoldenYAML(path.Join("testdata", "http", goldenFileName)))
				//
				// // then
				// Expect(snap.(*cache.Snapshot).Consistent()).To(Succeed())

				retry := &core_mesh.RetryResource{
					Meta: &test_model.ResourceMeta{
						Mesh: "default",
						Name: "web-to-backend",
					},
					Spec: &mesh_proto.Retry{
						Sources:      nil,
						Destinations: nil,
						Conf: &mesh_proto.Retry_Conf{
							Http: &mesh_proto.Retry_Conf_Http{
								NumRetries: util_proto.UInt32(7),
								RetriableMethods: []mesh_proto.HttpMethod{
									mesh_proto.HttpMethod_GET,
									mesh_proto.HttpMethod_POST,
								},
								PerTryTimeout: util_proto.Duration(time.Second * 1),
								BackOff: &mesh_proto.Retry_Conf_BackOff{
									BaseInterval: util_proto.Duration(time.Nanosecond * 200000000),
									MaxInterval:  util_proto.Duration(time.Nanosecond * 500000000),
								},
							},
						},
					},
				}

				configurers := gateway.RetryRouteConfigurers(
					core_mesh.ProtocolHTTP,
					retry,
				)

				builder := route.RouteBuilder{}
				resource, err := builder.Configure(configurers...).Build()
				Expect(err).To(BeNil())
				Expect(resource).To(BeNil())
			},
			entries,
		)
	})
})
