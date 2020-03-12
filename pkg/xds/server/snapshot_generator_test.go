package server

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/permissions"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	model "github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
	util_cache "github.com/Kong/kuma/pkg/util/cache"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/template"
)

var _ = Describe("Reconcile", func() {
	Describe("templateSnapshotGenerator", func() {

		gen := templateSnapshotGenerator{
			ProxyTemplateResolver: &simpleProxyTemplateResolver{
				ResourceManager:      manager.NewResourceManager(memory.NewStore()),
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
					Resource: &mesh_core.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
						Spec: mesh_proto.Mesh{
							Mtls: &mesh_proto.Mesh_Mtls{
								Enabled: true,
							},
						},
					},
				},
			}

			dataplane := mesh_proto.Dataplane{}
			dpBytes, err := ioutil.ReadFile(filepath.Join("testdata", "dataplane.input.yaml"))
			Expect(err).ToNot(HaveOccurred())
			Expect(util_proto.FromYAML(dpBytes, &dataplane)).To(Succeed())

			proxy := &model.Proxy{
				Id: model.ProxyId{Name: "demo.web1"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name:    "web1",
						Mesh:    "demo",
						Version: "1",
					},
					Spec: dataplane,
				},
				TrafficPermissions: permissions.MatchedPermissions{
					mesh_proto.InboundInterface{
						DataplaneIP:   "192.168.0.1",
						DataplanePort: 80,
						WorkloadPort:  8080,
					}: &mesh_core.TrafficPermissionResourceList{
						Items: []*mesh_core.TrafficPermissionResource{
							&mesh_core.TrafficPermissionResource{
								Meta: &test_model.ResourceMeta{
									Name: "tp-1",
									Mesh: "default",
								},
								Spec: mesh_proto.TrafficPermission{
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
				Metadata: &model.DataplaneMetadata{},
			}

			// when
			s, err := gen.GenerateSnapshot(ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := util_cache.ToDeltaDiscoveryResponse(s)
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(resp)
			// then
			Expect(err).ToNot(HaveOccurred())

			expected, err := ioutil.ReadFile(filepath.Join("testdata", "envoy-config.golden.yaml"))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(expected))
		})
	})
})
