package generator_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("DNSGenerator", func() {
	type testCase struct {
		dataplaneFile string
		expected      string
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// setup
			gen := &generator.DNSGenerator{}
			ctx := xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{},
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{},
					},
					VIPDomains: []xds_types.VIPDomains{
						{Address: "240.0.0.1", Domains: []string{"httpbin.mesh"}},
						{Address: "240.0.0.0", Domains: []string{"backend.test-ns.svc.8080.mesh", "backend_test-ns_svc_8080.mesh"}},
						{Address: "2001:db8::ff00:42:8329", Domains: []string{"frontend.test-ns.svc.8080.mesh", "frontend_test-ns_svc_8080.mesh"}}, // this is ignored because there is no outbounds for it
					},
				},
			}

			dataplane := mesh_proto.Dataplane{}
			dpBytes, err := os.ReadFile(filepath.Join("testdata", "dns", given.dataplaneFile))
			Expect(err).ToNot(HaveOccurred())
			Expect(util_proto.FromYAML(dpBytes, &dataplane)).To(Succeed())
			proxy := &model.Proxy{
				Id: *model.BuildProxyId("", "side-car"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "1",
					},
					Spec: &dataplane,
				},
				APIVersion: envoy_common.APIV3,
				Routing:    model.Routing{},
				Metadata: &model.DataplaneMetadata{
					DNSPort: 53001,
					Version: &mesh_proto.Version{Envoy: &mesh_proto.EnvoyVersion{Version: "1.20.0"}},
				},
			}

			for _, dppOutbound := range dataplane.GetNetworking().GetOutbound() {
				proxy.Outbounds = append(proxy.Outbounds, &xds_types.Outbound{LegacyOutbound: dppOutbound})
			}

			// when
			rs, err := gen.Generate(context.Background(), nil, ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and output matches golden files
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "dns", given.expected)))
		},
		Entry("01. DNS enabled", testCase{
			dataplaneFile: "1-dataplane.input.yaml",
			expected:      "1-envoy-config.golden.yaml",
		}),
		Entry("02. DNS disabled", testCase{
			dataplaneFile: "2-dataplane.input.yaml",
			expected:      "2-envoy-config.golden.yaml",
		}),
		Entry("03. DNS enabled no ipv6", testCase{
			dataplaneFile: "3-dataplane.input.yaml",
			expected:      "3-envoy-config.golden.yaml",
		}),
	)
})
