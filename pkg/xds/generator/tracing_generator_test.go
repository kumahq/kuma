package generator_test

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"

	"github.com/kumahq/kuma/pkg/xds/generator"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"

	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("TracingProxyGenerator", func() {

	type testCase struct {
		ctx          xds_context.Context
		proxy        *core_xds.Proxy
		expectedFile string
	}

	DescribeTable("should not generate Envoy xDS resources unless tracing is present",
		func(given testCase) {
			// setup
			gen := &generator.TracingProxyGenerator{}

			// when
			rs, err := gen.Generate(given.ctx, given.proxy)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(rs).To(BeNil())
		},
		Entry("Mesh has no Tracing configuration", testCase{
			proxy: &model.Proxy{
				Id: model.ProxyId{Name: "demo.backend-01"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
				},
				APIVersion: envoy_common.APIV2,
			},
		}),
	)

	DescribeTable("should generate Envoy xDS resources if tracing backend is present",
		func(given testCase) {
			// given
			gen := &generator.TracingProxyGenerator{}

			// when
			rs, err := gen.Generate(given.ctx, given.proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			resp, err := rs.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())

			expected, err := ioutil.ReadFile(filepath.Join("testdata", "tracing", given.expectedFile))
			Expect(err).ToNot(HaveOccurred())

			Expect(actual).To(MatchYAML(expected))
		},
		Entry("should create cluster for Zipkin", testCase{
			proxy: &model.Proxy{
				Id: model.ProxyId{Name: "demo.backend-01"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name: "backend-01",
						Mesh: "demo",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
						},
					},
				},
				APIVersion: envoy_common.APIV2,
				Policies: model.MatchedPolicies{
					TracingBackend: &mesh_proto.TracingBackend{
						Name: "zipkin",
						Type: mesh_proto.TracingZipkinType,
						Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
							Url: "http://zipkin.us:9090/v2/spans",
						}),
					},
				},
			},
			expectedFile: "zipkin.envoy-config.golden.yaml",
		}),
	)
})
