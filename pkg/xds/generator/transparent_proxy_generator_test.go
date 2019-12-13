package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	model "github.com/Kong/kuma/pkg/core/xds"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/generator"

	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("TransparentProxyGenerator", func() {

	type testCase struct {
		proxy    *model.Proxy
		expected string
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// setup
			gen := &generator.TransparentProxyGenerator{}
			ctx := xds_context.Context{}

			// when
			rs, err := gen.Generate(ctx, given.proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			// then
			resp := generator.ResourceList(rs).ToDeltaDiscoveryResponse()
			actual, err := util_proto.ToYAML(resp)

			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("transparent_proxying=false", testCase{
			proxy: &model.Proxy{
				Id: model.ProxyId{Name: "side-car"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "v1",
					},
				},
			},
			expected: `
        {}
`,
		}),
		Entry("transparent_proxying=true", testCase{
			proxy: &model.Proxy{
				Id: model.ProxyId{Name: "side-car"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "v1",
					},
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
								RedirectPort: 15001,
							},
						},
					},
				},
			},
			expected: `
        resources:
        - name: catch_all
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            address:
              socketAddress:
                address: 0.0.0.0
                portValue: 15001
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: pass_through
                  statPrefix: pass_through
            name: catch_all
            useOriginalDst: true
          version: v1
        - name: pass_through
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            connectTimeout: 5s
            lbPolicy: ORIGINAL_DST_LB
            name: pass_through
            type: ORIGINAL_DST
          version: v1
`,
		}),
	)
})
