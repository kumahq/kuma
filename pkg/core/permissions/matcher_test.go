package permissions_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/permissions"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Match", func() {

	type testCase struct {
		dataplane *core_mesh.DataplaneResource
		mesh      *core_mesh.MeshResource
		policies  []*core_mesh.TrafficPermissionResource
		expected  map[mesh_proto.InboundInterface]string
	}

	DescribeTable("should find best matched policy",
		func(given testCase) {
			manager := core_manager.NewResourceManager(memory.NewStore())
			matcher := permissions.TrafficPermissionsMatcher{ResourceManager: manager}

			err := manager.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			for _, p := range given.policies {
				err := manager.Create(context.Background(), p, store.CreateByKey(p.Meta.GetName(), "default"))
				Expect(err).ToNot(HaveOccurred())
			}

			bestMatched, err := matcher.Match(context.Background(), given.dataplane, given.mesh)
			Expect(err).ToNot(HaveOccurred())
			Expect(bestMatched).To(HaveLen(len(given.expected)))
			for iface, policy := range bestMatched {
				Expect(given.expected[iface]).To(Equal(policy.GetMeta().GetName()))
			}
		},
		Entry("2 inbounds dataplane with additional service, 2 policies", testCase{
			mesh: &core_mesh.MeshResource{
				Meta: &model.ResourceMeta{
					Name: "default",
				},
				Spec: &mesh_proto.Mesh{
					Metrics: &mesh_proto.Metrics{
						EnabledBackend: "prometheus-1",
						Backends: []*mesh_proto.MetricsBackend{
							{
								Name: "prometheus-1",
								Type: mesh_proto.MetricsPrometheusType,
								Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
									Port: 1234,
									Path: "/non-standard-path",
									Tags: map[string]string{
										"kuma.io/service": "dataplane-metrics",
									},
								}),
							},
						},
					},
				},
			},
			dataplane: &core_mesh.DataplaneResource{
				Meta: &model.ResourceMeta{
					Mesh: "default",
					Name: "dp1",
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        8080,
								ServicePort: 8081,
								Tags: map[string]string{
									"kuma.io/service":  "web",
									"version":          "0.1",
									"region":           "eu",
									"kuma.io/protocol": "http",
								},
							},
							{
								Port:        8081,
								ServicePort: 8082,
								Tags: map[string]string{
									"kuma.io/service":  "web-api",
									"version":          "0.1.2",
									"region":           "us",
									"kuma.io/protocol": "http",
								},
							},
						},
					},
				},
			},
			policies: []*core_mesh.TrafficPermissionResource{
				{
					Meta: &model.ResourceMeta{
						Mesh: "default",
						Name: "more-specific-kong-to-web",
					},
					Spec: &mesh_proto.TrafficPermission{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service": "kong",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service": "web",
									"version":         "0.1",
								},
							},
						},
					},
				},
				{
					Meta: &model.ResourceMeta{
						Mesh: "default",
						Name: "less-specific-kong-to-web",
					},
					Spec: &mesh_proto.TrafficPermission{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service": "kong",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service": "web",
								},
							},
						},
					},
				},
				{
					Meta: &model.ResourceMeta{
						Mesh: "default",
						Name: "metrics",
					},
					Spec: &mesh_proto.TrafficPermission{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service": "prometheus",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service": "dataplane-metrics",
								},
							},
						},
					},
				},
			},
			expected: map[mesh_proto.InboundInterface]string{
				{DataplaneAdvertisedIP: "192.168.0.1", DataplaneIP: "192.168.0.1", WorkloadIP: "127.0.0.1", WorkloadPort: 8081, DataplanePort: 8080}: "more-specific-kong-to-web",
				{DataplaneAdvertisedIP: "192.168.0.1", DataplaneIP: "192.168.0.1", WorkloadIP: "127.0.0.1", WorkloadPort: 1234, DataplanePort: 1234}: "metrics",
			},
		}),
	)

	Context("MatchExternalServices", func() {
		type testCase struct {
			dataplane        *core_mesh.DataplaneResource
			policies         []*core_mesh.TrafficPermissionResource
			externalServices []*core_mesh.ExternalServiceResource
			expected         map[string]bool
		}

		DescribeTable("should find the policy",
			func(given testCase) {
				manager := core_manager.NewResourceManager(memory.NewStore())

				err := manager.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
				Expect(err).ToNot(HaveOccurred())

				for _, p := range given.policies {
					err := manager.Create(context.Background(), p, store.CreateByKey(p.Meta.GetName(), "default"))
					Expect(err).ToNot(HaveOccurred())
				}

				es := &core_mesh.ExternalServiceResourceList{
					Items: given.externalServices,
				}
				tp := &core_mesh.TrafficPermissionResourceList{
					Items: given.policies,
				}
				matchedEs, err := permissions.MatchExternalServices(given.dataplane, es, tp)
				Expect(err).ToNot(HaveOccurred())

				Expect(given.expected).To(HaveLen(len(matchedEs)))
				for _, externalService := range matchedEs {
					Expect(given.expected[externalService.GetMeta().GetName()]).To(BeTrue())
				}
			},
			Entry("should match external services that matches traffic permission", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Meta: &model.ResourceMeta{
						Mesh: "default",
						Name: "dp1",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Port:        8080,
									ServicePort: 8081,
									Tags: map[string]string{
										"kuma.io/service": "web",
									},
								},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Port: 8080,
									Tags: map[string]string{
										"kuma.io/service": "httpbin",
									},
								},
							},
						},
					},
				},
				externalServices: []*core_mesh.ExternalServiceResource{
					{
						Meta: &model.ResourceMeta{
							Mesh: "default",
							Name: "httpbin",
						},
						Spec: &mesh_proto.ExternalService{
							Tags: map[string]string{
								"kuma.io/service": "httpbin",
							},
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org",
							},
						},
					},
					{ // this won't be matched since there is no traffic permission for it
						Meta: &model.ResourceMeta{
							Mesh: "default",
							Name: "google",
						},
						Spec: &mesh_proto.ExternalService{
							Tags: map[string]string{
								"kuma.io/service": "google",
							},
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "google.com",
							},
						},
					},
				},
				policies: []*core_mesh.TrafficPermissionResource{
					{
						Meta: &model.ResourceMeta{
							Mesh: "default",
							Name: "web-to-httpbin",
						},
						Spec: &mesh_proto.TrafficPermission{
							Sources: []*mesh_proto.Selector{
								{
									Match: map[string]string{
										"kuma.io/service": "web",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: map[string]string{
										"kuma.io/service": "httpbin",
									},
								},
							},
						},
					},
				},
				expected: map[string]bool{
					"httpbin": true,
				},
			}),
			Entry("should match all external services because of the traffic permission that matches all", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Meta: &model.ResourceMeta{
						Mesh: "default",
						Name: "dp1",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "192.168.0.1",
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Port:        8080,
									ServicePort: 8081,
									Tags: map[string]string{
										"kuma.io/service": "web",
									},
								},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Port: 8080,
									Tags: map[string]string{
										"kuma.io/service": "httpbin",
									},
								},
							},
						},
					},
				},
				externalServices: []*core_mesh.ExternalServiceResource{
					{
						Meta: &model.ResourceMeta{
							Mesh: "default",
							Name: "httpbin",
						},
						Spec: &mesh_proto.ExternalService{
							Tags: map[string]string{
								"kuma.io/service": "httpbin",
							},
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org",
							},
						},
					},
					{ // this won't be matched since there is no traffic permission for it
						Meta: &model.ResourceMeta{
							Mesh: "default",
							Name: "google",
						},
						Spec: &mesh_proto.ExternalService{
							Tags: map[string]string{
								"kuma.io/service": "google",
							},
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "google.com",
							},
						},
					},
				},
				policies: []*core_mesh.TrafficPermissionResource{
					{
						Meta: &model.ResourceMeta{
							Mesh: "default",
							Name: "all",
						},
						Spec: &mesh_proto.TrafficPermission{
							Sources: []*mesh_proto.Selector{
								{
									Match: map[string]string{
										"kuma.io/service": "*",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: map[string]string{
										"kuma.io/service": "*",
									},
								},
							},
						},
					},
				},
				expected: map[string]bool{
					"httpbin": true,
					"google":  true,
				},
			}),
		)
	})
})
