package permissions_test

import (
	"context"

	util_proto "github.com/Kong/kuma/pkg/util/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/permissions"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/test/resources/model"
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

			err := manager.Create(context.Background(), &core_mesh.MeshResource{}, store.CreateByKey("default", "default"))
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
					Mesh: "default",
					Name: "default",
				},
				Spec: mesh_proto.Mesh{
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
										"service": "dataplane-metrics",
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
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        8080,
								ServicePort: 8081,
								Tags: map[string]string{
									"service":  "web",
									"version":  "0.1",
									"region":   "eu",
									"protocol": "http",
								},
							},
							{
								Port:        8081,
								ServicePort: 8082,
								Tags: map[string]string{
									"service":  "web-api",
									"version":  "0.1.2",
									"region":   "us",
									"protocol": "http",
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
					Spec: mesh_proto.TrafficPermission{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "kong",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "web",
									"version": "0.1",
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
					Spec: mesh_proto.TrafficPermission{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "kong",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "web",
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
					Spec: mesh_proto.TrafficPermission{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "prometheus",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "dataplane-metrics",
								},
							},
						},
					},
				},
			},
			expected: map[mesh_proto.InboundInterface]string{
				mesh_proto.InboundInterface{DataplaneIP: "192.168.0.1", WorkloadPort: 8081, DataplanePort: 8080}: "more-specific-kong-to-web",
				mesh_proto.InboundInterface{DataplaneIP: "192.168.0.1", WorkloadPort: 1234, DataplanePort: 1234}: "metrics",
			},
		}),
	)
})
