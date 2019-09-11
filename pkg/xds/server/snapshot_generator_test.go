package server

import (
	"github.com/Kong/kuma/pkg/core/permissions"
	"io/ioutil"
	"path/filepath"

	"github.com/Kong/kuma/pkg/core/resources/manager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	model "github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	util_cache "github.com/Kong/kuma/pkg/util/cache"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/template"

	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("Reconcile", func() {
	Describe("templateSnapshotGenerator", func() {

		gen := templateSnapshotGenerator{
			ProxyTemplateResolver: &simpleProxyTemplateResolver{
				ResourceManager:      manager.NewResourceManager(memory.NewStore()),
				DefaultProxyTemplate: template.DefaultProxyTemplate,
			},
		}

		type testCase struct {
			dataplaneFile   string
			envoyConfigFile string
		}

		DescribeTable("Generate Snapshot per Envoy Node",
			func(given testCase) {
				// given
				ctx := xds_context.Context{
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsLocation: "kuma-system:5677",
						SdsTlsCert:  []byte("12345"),
					},
					Mesh: xds_context.MeshContext{
						TlsEnabled: true,
					},
				}

				dataplane := mesh_proto.Dataplane{}
				dpBytes, err := ioutil.ReadFile(filepath.Join("testdata", given.dataplaneFile))
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
					TrafficPermissions: permissions.MatchedPermissions{
						"192.168.0.1:80:8080": &mesh_core.TrafficPermissionResourceList{
							Items: []*mesh_core.TrafficPermissionResource{
								&mesh_core.TrafficPermissionResource{
									Meta: &test_model.ResourceMeta{
										Name:      "tp-1",
										Mesh:      "default",
										Namespace: "default",
									},
									Spec: mesh_proto.TrafficPermission{
										Rules: []*mesh_proto.TrafficPermission_Rule{
											{
												Sources: []*mesh_proto.TrafficPermission_Rule_Selector{
													{
														Match: map[string]string{
															"service": "web1",
															"version": "1.0",
														},
													},
												},
												Destinations: []*mesh_proto.TrafficPermission_Rule_Selector{
													{
														Match: map[string]string{
															"service": "backend1",
															"env":     "dev",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}

				// when
				s, err := gen.GenerateSnapshot(ctx, proxy)

				// then
				Expect(err).ToNot(HaveOccurred())

				// then
				resp := util_cache.ToDeltaDiscoveryResponse(s)
				actual, err := util_proto.ToYAML(resp)
				Expect(err).ToNot(HaveOccurred())

				expected, err := ioutil.ReadFile(filepath.Join("testdata", given.envoyConfigFile))
				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(expected))
			},
			//Entry("01. transparent_proxying=false, ip_addresses=0, ports=0", testCase{
			//	dataplaneFile:   "1-dataplane.input.yaml",
			//	envoyConfigFile: "1-envoy-config.golden.yaml",
			//}),
			//Entry("02. transparent_proxying=true, ip_addresses=0, ports=0", testCase{
			//	dataplaneFile:   "2-dataplane.input.yaml",
			//	envoyConfigFile: "2-envoy-config.golden.yaml",
			//}),
			//Entry("03. transparent_proxying=false, ip_addresses=1, ports=1", testCase{
			//	dataplaneFile:   "3-dataplane.input.yaml",
			//	envoyConfigFile: "3-envoy-config.golden.yaml",
			//}),
			//Entry("04. transparent_proxying=true, ip_addresses=1, ports=1", testCase{
			//	dataplaneFile:   "4-dataplane.input.yaml",
			//	envoyConfigFile: "4-envoy-config.golden.yaml",
			//}),
			//Entry("05. transparent_proxying=false, ip_addresses=1, ports=2", testCase{
			//	dataplaneFile:   "5-dataplane.input.yaml",
			//	envoyConfigFile: "5-envoy-config.golden.yaml",
			//}),
			//Entry("06. transparent_proxying=true, ip_addresses=1, ports=2", testCase{
			//	dataplaneFile:   "6-dataplane.input.yaml",
			//	envoyConfigFile: "6-envoy-config.golden.yaml",
			//}),
			//Entry("07. transparent_proxying=false, ip_addresses=2, ports=2", testCase{
			//	dataplaneFile:   "7-dataplane.input.yaml",
			//	envoyConfigFile: "7-envoy-config.golden.yaml",
			//}),
			Entry("08. 	transparent_proxying=true, ip_addresses=2, ports=2", testCase{
				dataplaneFile:   "8-dataplane.input.yaml",
				envoyConfigFile: "8-envoy-config.golden.yaml",
			}),
		)
	})
})
