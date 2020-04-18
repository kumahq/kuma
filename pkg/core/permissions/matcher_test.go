package permissions_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/permissions"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("Match", func() {

	type testCase struct {
		dataplane *mesh.DataplaneResource
		policies  []*mesh.TrafficPermissionResource
		expected  map[mesh_proto.InboundInterface]string
	}

	DescribeTable("should find best matched policy",
		func(given testCase) {
			manager := core_manager.NewResourceManager(memory.NewStore())
			matcher := permissions.TrafficPermissionsMatcher{ResourceManager: manager}

			err := manager.Create(context.Background(), &mesh.MeshResource{}, store.CreateByKey("default", "default"))
			Expect(err).ToNot(HaveOccurred())

			for _, p := range given.policies {
				err := manager.Create(context.Background(), p, store.CreateByKey(p.Meta.GetName(), "default"))
				Expect(err).ToNot(HaveOccurred())
			}

			bestMatched, err := matcher.Match(context.Background(), given.dataplane)
			Expect(err).ToNot(HaveOccurred())
			Expect(bestMatched).To(HaveLen(len(given.expected)))
			for iface, policy := range bestMatched {
				Expect(given.expected[iface]).To(Equal(policy.GetMeta().GetName()))
			}
		},
		Entry("1 inbound dataplane, 2 policies", testCase{
			dataplane: &mesh.DataplaneResource{
				Meta: &model.ResourceMeta{
					Mesh: "default",
					Name: "dp1",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								ServicePort: 8080,
								Tags: map[string]string{
									"service":  "web",
									"version":  "0.1",
									"region":   "eu",
									"protocol": "http",
								},
							},
							{
								ServicePort: 8081,
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
			policies: []*mesh.TrafficPermissionResource{
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
			},
			expected: map[mesh_proto.InboundInterface]string{
				mesh_proto.InboundInterface{WorkloadPort: 8080}: "more-specific-kong-to-web",
			},
		}),
	)
})
