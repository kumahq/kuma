package server

import (
	"github.com/Kong/kuma/pkg/core/logs"
	"github.com/Kong/kuma/pkg/core/permissions"
	test_resources "github.com/Kong/kuma/pkg/test/resources"
	"io/ioutil"
	"path/filepath"

	"github.com/Kong/kuma/pkg/core/resources/manager"
	. "github.com/onsi/ginkgo"
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
				ResourceManager:      manager.NewResourceManager(memory.NewStore(), test_resources.Global()),
				DefaultProxyTemplate: template.DefaultProxyTemplate,
			},
		}

		It("Generate Snapshot per Envoy Node", func() {
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
			dpBytes, err := ioutil.ReadFile(filepath.Join("testdata", "dataplane.input.yaml"))
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
											Sources: []*mesh_proto.Selector{
												{
													Match: map[string]string{
														"service": "web1",
														"version": "1.0",
													},
												},
											},
											Destinations: []*mesh_proto.Selector{
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
				Logs: logs.NewMatchedLogs(),
			}

			// when
			s, err := gen.GenerateSnapshot(ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			// then
			resp := util_cache.ToDeltaDiscoveryResponse(s)
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())

			expected, err := ioutil.ReadFile(filepath.Join("testdata", "envoy-config.golden.yaml"))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(expected))
		})
	})
})
