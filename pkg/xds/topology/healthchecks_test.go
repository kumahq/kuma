package topology_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/topology"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"

	"github.com/golang/protobuf/ptypes"
)

var _ = Describe("HealthCheck", func() {

	var ctx context.Context
	var rm core_manager.ResourceManager

	BeforeEach(func() {
		ctx = context.Background()
		rm = core_manager.NewResourceManager(memory_resources.NewStore())
	})

	Describe("GetHealthChecks()", func() {

		It("should pick the best matching HealthCheck for each destination service", func() {
			// given
			mesh := &mesh_core.MeshResource{ // mesh that is relevant to this test case
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "demo",
				},
			}
			otherMesh := &mesh_core.MeshResource{ // mesh that is irrelevant to this test case
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "default",
				},
			}
			backend := &mesh_core.DataplaneResource{ // dataplane that is a source of traffic
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "backend",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags:        map[string]string{"service": "backend", "region": "eu"},
								Port:        8080,
								ServicePort: 18080,
							},
							{
								Tags:        map[string]string{"service": "frontend", "region": "eu"},
								Port:        7070,
								ServicePort: 17070,
							},
						},
						Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
							{Service: "redis", Port: 10001},
							{Service: "elastic", Port: 10002},
						},
					},
				},
			}
			destinations := core_xds.DestinationMap{
				"redis":   core_xds.TagSelectorSet{mesh_proto.MatchService("redis")},
				"elastic": core_xds.TagSelectorSet{mesh_proto.MatchService("elastic")},
			}
			healthCheckRedis := &mesh_core.HealthCheckResource{ // health checks for `redis` service
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "healthcheck-redis",
				},
				Spec: mesh_proto.HealthCheck{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"service": "frontend"}},
						{Match: mesh_proto.TagSelector{"service": "backend"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"service": "redis"}},
					},
					Conf: &mesh_proto.HealthCheck_Conf{
						ActiveChecks: &mesh_proto.HealthCheck_Conf_Active{
							Interval:           ptypes.DurationProto(5 * time.Second),
							Timeout:            ptypes.DurationProto(4 * time.Second),
							UnhealthyThreshold: 3,
							HealthyThreshold:   2,
						},
					},
				},
			}
			healthCheckElastic := &mesh_core.HealthCheckResource{ // health checks for `elastic` service
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "healthcheck-elastic",
				},
				Spec: mesh_proto.HealthCheck{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"service": "*"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"service": "elastic"}},
					},
					Conf: &mesh_proto.HealthCheck_Conf{
						PassiveChecks: &mesh_proto.HealthCheck_Conf_Passive{
							UnhealthyThreshold: 1,
							PenaltyInterval:    ptypes.DurationProto(6 * time.Second),
						},
					},
				},
			}
			healthCheckEverything := &mesh_core.HealthCheckResource{ // health checks for any service
				Meta: &test_model.ResourceMeta{
					Mesh: "default", // other mesh
					Name: "healthcheck-everything",
				},
				Spec: mesh_proto.HealthCheck{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"service": "*"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"service": "*"}},
					},
					Conf: &mesh_proto.HealthCheck_Conf{
						PassiveChecks: &mesh_proto.HealthCheck_Conf_Passive{
							UnhealthyThreshold: 20,
							PenaltyInterval:    ptypes.DurationProto(30 * time.Second),
						},
					},
				},
			}
			for _, resource := range []core_model.Resource{mesh, backend, healthCheckRedis, healthCheckElastic, otherMesh, healthCheckEverything} {
				// when
				err := rm.Create(ctx, resource, core_store.CreateBy(core_model.MetaToResourceKey(resource.GetMeta())))
				// then
				Expect(err).ToNot(HaveOccurred())
			}

			// when
			healthChecks, err := GetHealthChecks(ctx, backend, destinations, rm)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(healthChecks).To(HaveLen(2))
			// and
			Expect(healthChecks).To(HaveKey("redis"))
			Expect(healthChecks["redis"].Meta.GetName()).To(Equal(healthCheckRedis.Meta.GetName()))
			Expect(healthChecks["redis"].Spec).To(Equal(healthCheckRedis.Spec))
			// and
			Expect(healthChecks).To(HaveKey("elastic"))
			Expect(healthChecks["elastic"].Meta.GetName()).To(Equal(healthCheckElastic.Meta.GetName()))
			Expect(healthChecks["elastic"].Spec).To(Equal(healthCheckElastic.Spec))
		})
	})

	Describe("BuildHealthCheckMap()", func() {
		sameMeta := func(meta1, meta2 core_model.ResourceMeta) bool {
			return meta1.GetMesh() == meta2.GetMesh() &&
				meta1.GetName() == meta2.GetName() &&
				meta1.GetVersion() == meta2.GetVersion()
		}
		type testCase struct {
			dataplane    *mesh_core.DataplaneResource
			destinations core_xds.DestinationMap
			healthChecks []*mesh_core.HealthCheckResource
			expected     core_xds.HealthCheckMap
		}
		DescribeTable("should correctly pick a single the most specific HealthCheck for each outbound interface",
			func(given testCase) {
				// setup
				expectedHealthChecks := core_xds.HealthCheckMap{}
				for service, expectedHealthCheck := range given.expected {
					for _, healthCheck := range given.healthChecks {
						if sameMeta(expectedHealthCheck.GetMeta(), healthCheck.GetMeta()) {
							expectedHealthChecks[service] = healthCheck
							break
						}
					}
					if _, ok := expectedHealthChecks[service]; !ok {
						expectedHealthChecks[service] = expectedHealthCheck
					}
				}
				if len(expectedHealthChecks) == 0 {
					expectedHealthChecks = nil
				}
				// when
				healthChecks := BuildHealthCheckMap(given.dataplane, given.destinations, given.healthChecks)
				// expect
				Expect(healthChecks).Should(Equal(expectedHealthChecks))
			},
			Entry("Dataplane without outbound interfaces (and therefore no destinations)", testCase{
				dataplane:    &mesh_core.DataplaneResource{},
				destinations: nil,
				healthChecks: nil,
				expected:     nil,
			}),
			Entry("if a destination service has no matching HealthChecks, none should be used", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"service": "backend"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis"},
								{Service: "elastic"},
							},
						},
					},
				},
				destinations: core_xds.DestinationMap{
					"redis":   core_xds.TagSelectorSet{mesh_proto.MatchService("redis")},
					"elastic": core_xds.TagSelectorSet{mesh_proto.MatchService("elastic")},
				},
				healthChecks: nil,
				expected:     nil,
			}),
			Entry("due to TrafficRoutes, a Dataplane might have more destinations than outbound interfaces", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"service": "backend"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis"},
							},
						},
					},
				},
				destinations: core_xds.DestinationMap{
					"redis":   core_xds.TagSelectorSet{mesh_proto.MatchService("redis")},
					"elastic": core_xds.TagSelectorSet{mesh_proto.MatchService("elastic")},
				},
				healthChecks: []*mesh_core.HealthCheckResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "healthcheck-elastic",
						},
						Spec: mesh_proto.HealthCheck{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "elastic"}},
							},
							Conf: &mesh_proto.HealthCheck_Conf{
								ActiveChecks: &mesh_proto.HealthCheck_Conf_Active{
									Interval:           ptypes.DurationProto(5 * time.Second),
									Timeout:            ptypes.DurationProto(4 * time.Second),
									UnhealthyThreshold: 3,
									HealthyThreshold:   2,
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "healthcheck-redis",
						},
						Spec: mesh_proto.HealthCheck{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "redis"}},
							},
							Conf: &mesh_proto.HealthCheck_Conf{
								PassiveChecks: &mesh_proto.HealthCheck_Conf_Passive{
									UnhealthyThreshold: 20,
									PenaltyInterval:    ptypes.DurationProto(30 * time.Second),
								},
							},
						},
					},
				},
				expected: core_xds.HealthCheckMap{
					"redis": &mesh_core.HealthCheckResource{
						Meta: &test_model.ResourceMeta{
							Name: "healthcheck-redis",
						},
					},
					"elastic": &mesh_core.HealthCheckResource{
						Meta: &test_model.ResourceMeta{
							Name: "healthcheck-elastic",
						},
					},
				},
			}),
			Entry("HealthChecks should be picked by latest creation time given two equally specific HealthChecks", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"service": "backend"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis"},
							},
						},
					},
				},
				destinations: core_xds.DestinationMap{
					"redis": core_xds.TagSelectorSet{mesh_proto.MatchService("redis")},
				},
				healthChecks: []*mesh_core.HealthCheckResource{
					{
						Meta: &test_model.ResourceMeta{
							Name:         "healthcheck-everything-passive",
							CreationTime: time.Unix(1, 1),
						},
						Spec: mesh_proto.HealthCheck{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: &mesh_proto.HealthCheck_Conf{
								ActiveChecks: &mesh_proto.HealthCheck_Conf_Active{
									Interval:           ptypes.DurationProto(5 * time.Second),
									Timeout:            ptypes.DurationProto(4 * time.Second),
									UnhealthyThreshold: 3,
									HealthyThreshold:   2,
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name:         "healthcheck-everything-active",
							CreationTime: time.Unix(0, 0),
						},
						Spec: mesh_proto.HealthCheck{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: &mesh_proto.HealthCheck_Conf{
								PassiveChecks: &mesh_proto.HealthCheck_Conf_Passive{
									UnhealthyThreshold: 20,
									PenaltyInterval:    ptypes.DurationProto(30 * time.Second),
								},
							},
						},
					},
				},
				expected: core_xds.HealthCheckMap{
					"redis": &mesh_core.HealthCheckResource{
						Meta: &test_model.ResourceMeta{
							Name: "healthcheck-everything-passive",
						},
					},
				},
			}),
			Entry("HealthCheck with a `source` selector by 2 tags should win over a HealthCheck with a `source` selector by 1 tag", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"service": "backend", "region": "eu"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis"},
							},
						},
					},
				},
				destinations: core_xds.DestinationMap{
					"redis": core_xds.TagSelectorSet{mesh_proto.MatchService("redis")},
				},
				healthChecks: []*mesh_core.HealthCheckResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "less-specific",
						},
						Spec: mesh_proto.HealthCheck{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "backend"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: &mesh_proto.HealthCheck_Conf{
								ActiveChecks: &mesh_proto.HealthCheck_Conf_Active{
									Interval:           ptypes.DurationProto(5 * time.Second),
									Timeout:            ptypes.DurationProto(4 * time.Second),
									UnhealthyThreshold: 3,
									HealthyThreshold:   2,
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
						Spec: mesh_proto.HealthCheck{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "backend", "region": "eu"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: &mesh_proto.HealthCheck_Conf{
								PassiveChecks: &mesh_proto.HealthCheck_Conf_Passive{
									UnhealthyThreshold: 20,
									PenaltyInterval:    ptypes.DurationProto(30 * time.Second),
								},
							},
						},
					},
				},
				expected: core_xds.HealthCheckMap{
					"redis": &mesh_core.HealthCheckResource{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
					},
				},
			}),
			Entry("HealthCheck with a `source` selector by an exact value should win over a HealthCheck with a `source` selector by a wildcard value", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"service": "backend", "region": "eu"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis"},
							},
						},
					},
				},
				destinations: core_xds.DestinationMap{
					"redis": core_xds.TagSelectorSet{mesh_proto.MatchService("redis")},
				},
				healthChecks: []*mesh_core.HealthCheckResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "less-specific",
						},
						Spec: mesh_proto.HealthCheck{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: &mesh_proto.HealthCheck_Conf{
								ActiveChecks: &mesh_proto.HealthCheck_Conf_Active{
									Interval:           ptypes.DurationProto(5 * time.Second),
									Timeout:            ptypes.DurationProto(4 * time.Second),
									UnhealthyThreshold: 3,
									HealthyThreshold:   2,
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
						Spec: mesh_proto.HealthCheck{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "backend"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: &mesh_proto.HealthCheck_Conf{
								PassiveChecks: &mesh_proto.HealthCheck_Conf_Passive{
									UnhealthyThreshold: 20,
									PenaltyInterval:    ptypes.DurationProto(30 * time.Second),
								},
							},
						},
					},
				},
				expected: core_xds.HealthCheckMap{
					"redis": &mesh_core.HealthCheckResource{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
					},
				},
			}),
			Entry("HealthCheck with a `destination` selector by an exact value should win over a HealthCheck with a `destination` selector by a wildcard value", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"service": "backend", "region": "eu"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis"},
							},
						},
					},
				},
				destinations: core_xds.DestinationMap{
					"redis": core_xds.TagSelectorSet{mesh_proto.MatchService("redis")},
				},
				healthChecks: []*mesh_core.HealthCheckResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "less-specific",
						},
						Spec: mesh_proto.HealthCheck{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: &mesh_proto.HealthCheck_Conf{
								ActiveChecks: &mesh_proto.HealthCheck_Conf_Active{
									Interval:           ptypes.DurationProto(5 * time.Second),
									Timeout:            ptypes.DurationProto(4 * time.Second),
									UnhealthyThreshold: 3,
									HealthyThreshold:   2,
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
						Spec: mesh_proto.HealthCheck{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "redis"}},
							},
							Conf: &mesh_proto.HealthCheck_Conf{
								PassiveChecks: &mesh_proto.HealthCheck_Conf_Passive{
									UnhealthyThreshold: 20,
									PenaltyInterval:    ptypes.DurationProto(30 * time.Second),
								},
							},
						},
					},
				},
				expected: core_xds.HealthCheckMap{
					"redis": &mesh_core.HealthCheckResource{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
					},
				},
			}),
			Entry("in case if HealthChecks have equal aggregate ranks, most specific one should be selected based on last creation time", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"service": "backend", "region": "eu"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis"},
							},
						},
					},
				},
				destinations: core_xds.DestinationMap{
					"redis": core_xds.TagSelectorSet{mesh_proto.MatchService("redis")},
				},
				healthChecks: []*mesh_core.HealthCheckResource{
					{
						Meta: &test_model.ResourceMeta{
							Name:         "equally-specific-2",
							CreationTime: time.Unix(1, 1),
						},
						Spec: mesh_proto.HealthCheck{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "redis"}},
							},
							Conf: &mesh_proto.HealthCheck_Conf{
								ActiveChecks: &mesh_proto.HealthCheck_Conf_Active{
									Interval:           ptypes.DurationProto(5 * time.Second),
									Timeout:            ptypes.DurationProto(4 * time.Second),
									UnhealthyThreshold: 3,
									HealthyThreshold:   2,
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name:         "equally-specific-1",
							CreationTime: time.Unix(0, 0),
						},
						Spec: mesh_proto.HealthCheck{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "backend"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: &mesh_proto.HealthCheck_Conf{
								PassiveChecks: &mesh_proto.HealthCheck_Conf_Passive{
									UnhealthyThreshold: 20,
									PenaltyInterval:    ptypes.DurationProto(30 * time.Second),
								},
							},
						},
					},
				},
				expected: core_xds.HealthCheckMap{
					"redis": &mesh_core.HealthCheckResource{
						Meta: &test_model.ResourceMeta{
							Name: "equally-specific-2",
						},
					},
				},
			}),
		)
	})

})
