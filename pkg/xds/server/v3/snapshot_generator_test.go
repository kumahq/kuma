package v3

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	model "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/xds"
	util_cache_v3 "github.com/kumahq/kuma/pkg/util/cache/v3"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/template"
)

var _ = Describe("Reconcile", func() {
	Describe("templateSnapshotGenerator", func() {

		gen := templateSnapshotGenerator{
			ProxyTemplateResolver: template.SequentialResolver(
				&template.SimpleProxyTemplateResolver{
					ReadOnlyResourceManager: manager.NewResourceManager(memory.NewStore()),
				},
				generator.DefaultTemplateResolver,
			),
		}

		It("Generate Snapshot per Envoy Node", func() {
			// given
			ctx := xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					Secrets: &xds.TestSecrets{},
				},
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
						Spec: &mesh_proto.Mesh{
							Mtls: &mesh_proto.Mesh_Mtls{
								EnabledBackend: "builtin",
								Backends: []*mesh_proto.CertificateAuthorityBackend{
									{
										Name: "builtin",
										Type: "builtin",
									},
								},
							},
						},
					},
				},
			}

			dataplane := mesh_proto.Dataplane{}
			dpBytes, err := os.ReadFile(filepath.Join("testdata", "dataplane.input.yaml"))
			Expect(err).ToNot(HaveOccurred())
			Expect(util_proto.FromYAML(dpBytes, &dataplane)).To(Succeed())

			proxy := &model.Proxy{
				Id:         *model.BuildProxyId("", "demo.web1"),
				APIVersion: envoy_common.APIV3,
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name:    "web1",
						Mesh:    "demo",
						Version: "1",
					},
					Spec: &dataplane,
				},
				Policies: model.MatchedPolicies{
					TrafficPermissions: model.TrafficPermissionMap{
						mesh_proto.InboundInterface{
							DataplaneAdvertisedIP: "192.168.0.1",
							DataplaneIP:           "192.168.0.1",
							DataplanePort:         80,
							WorkloadIP:            "127.0.0.1",
							WorkloadPort:          8080,
						}: &core_mesh.TrafficPermissionResource{
							Meta: &test_model.ResourceMeta{
								Name: "tp-1",
								Mesh: "default",
							},
							Spec: &mesh_proto.TrafficPermission{
								Sources: []*mesh_proto.Selector{
									{
										Match: map[string]string{
											"kuma.io/service": "web1",
											"version":         "1.0",
										},
									},
								},
								Destinations: []*mesh_proto.Selector{
									{
										Match: map[string]string{
											"kuma.io/service": "backend1",
											"env":             "dev",
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
			resp, err := util_cache_v3.ToDeltaDiscoveryResponse(s)
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(resp)
			// then
			Expect(err).ToNot(HaveOccurred())

			Expect(actual).To(matchers.MatchGoldenYAML(filepath.Join("testdata", "envoy-config.golden.yaml")))
		})
	})
})
