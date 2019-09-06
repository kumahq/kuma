package generator_test

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/xds"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"path/filepath"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
	xds_envoy "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/envoy"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/generator"

	test_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/model"
)

var _ = Describe("InboundProxyGenerator", func() {

	type testCase struct {
		dataplaneFile   string
		envoyConfigFile string
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// setup
			gen := &generator.InboundProxyGenerator{}
			ctx := xds_envoy.Context{
				ControlPlane: &xds_envoy.ControlPlaneContext{
					Config: xds.SnapshotConfig{
						SdsLocation: "konvoy-system:5677",
					},
					SdsTlsCert: []byte("12345"),
				},
			}

			dataplane := mesh_proto.Dataplane{}
			dpBytes, err := ioutil.ReadFile(filepath.Join("testdata", "inbound-proxy", given.dataplaneFile))
			Expect(err).ToNot(HaveOccurred())
			Expect(util_proto.FromYAML(dpBytes, &dataplane)).To(Succeed())
			proxy := &model.Proxy{
				Id: model.ProxyId{Name: "side-car", Namespace: "default"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "1",
					},
					Spec: dataplane,
				},
			}

			// when
			rs, err := gen.Generate(ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			// then
			resp := generator.ResourceList(rs).ToDeltaDiscoveryResponse()
			actual, err := util_proto.ToYAML(resp)

			expected, err := ioutil.ReadFile(filepath.Join("testdata", "inbound-proxy", given.envoyConfigFile))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(expected))
		},
		Entry("transparent_proxying=false, ip_addresses=0, ports=0", testCase{
			dataplaneFile:   "1-dataplane.yaml",
			envoyConfigFile: "1-envoy-config.yaml",
		}),
		Entry("transparent_proxying=true, ip_addresses=0, ports=0", testCase{
			dataplaneFile:   "2-dataplane.yaml",
			envoyConfigFile: "1-envoy-config.yaml",
		}),
		Entry("transparent_proxying=false, ip_addresses=1, ports=1", testCase{
			dataplaneFile:   "3-dataplane.yaml",
			envoyConfigFile: "3-envoy-config.yaml",
		}),
		Entry("transparent_proxying=true, ip_addresses=1, ports=1", testCase{
			dataplaneFile:   "4-dataplane.yaml",
			envoyConfigFile: "4-envoy-config.yaml",
		}),
		Entry("transparent_proxying=false, ip_addresses=1, ports=2", testCase{
			dataplaneFile:   "5-dataplane.yaml",
			envoyConfigFile: "5-envoy-config.yaml",
		}),
		Entry("transparent_proxying=true, ip_addresses=1, ports=2", testCase{
			dataplaneFile:   "6-dataplane.yaml",
			envoyConfigFile: "6-envoy-config.yaml",
		}),
		Entry("transparent_proxying=false, ip_addresses=2, ports=2", testCase{
			dataplaneFile:   "7-dataplane.yaml",
			envoyConfigFile: "7-envoy-config.yaml",
		}),
		Entry("transparent_proxying=true, ip_addresses=2, ports=2", testCase{
			dataplaneFile:   "8-dataplane.yaml",
			envoyConfigFile: "8-envoy-config.yaml",
		}),
	)
})
