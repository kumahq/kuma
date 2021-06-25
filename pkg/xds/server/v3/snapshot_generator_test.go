package v3

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/matchers"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	model "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_cache_v3 "github.com/kumahq/kuma/pkg/util/cache/v3"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/template"
)

var _ = Describe("Reconcile", func() {
	Describe("templateSnapshotGenerator", func() {

		gen := templateSnapshotGenerator{
			ProxyTemplateResolver: &template.SimpleProxyTemplateResolver{
				ReadOnlyResourceManager: manager.NewResourceManager(memory.NewStore()),
				DefaultProxyTemplate:    template.DefaultProxyTemplate,
			},
		}

		It("Generate Snapshot per Envoy Node", func() {
			// given
			ctx := xds_context.Context{
				ConnectionInfo: xds_context.ConnectionInfo{
					Authority: "kuma-system:5677",
				},
				ControlPlane: &xds_context.ControlPlaneContext{
					SdsTlsCert: []byte("12345"),
				},
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
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
			dpBytes, err := ioutil.ReadFile(filepath.Join("testdata", "dataplane.input.yaml"))
			Expect(err).ToNot(HaveOccurred())
			Expect(util_proto.FromYAML(dpBytes, &dataplane)).To(Succeed())

			proxy := &model.Proxy{
				Id:         *model.BuildProxyId("", "demo.web1"),
				APIVersion: envoy_common.APIV3,
				Dataplane: &mesh_core.DataplaneResource{
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
						}: &mesh_core.TrafficPermissionResource{
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
