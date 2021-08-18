package ratelimits_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	. "github.com/kumahq/kuma/pkg/core/ratelimits"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Match", func() {
	dataplaneWithInboundsFunc := func(inbounds []*mesh_proto.Dataplane_Networking_Inbound) *mesh.DataplaneResource {
		return &mesh.DataplaneResource{
			Meta: &model.ResourceMeta{
				Mesh: "default",
				Name: "dp1",
			},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Inbound: inbounds,
				},
			},
		}
	}

	policyWithDestinationsFunc := func(name string, creationTime time.Time, sources, destinations []*mesh_proto.Selector) *mesh.RateLimitResource {
		return &mesh.RateLimitResource{
			Meta: &model.ResourceMeta{
				Name:         name,
				CreationTime: creationTime,
			},
			Spec: &mesh_proto.RateLimit{
				Sources:      sources,
				Destinations: destinations,
				Conf: &mesh_proto.RateLimit_Conf{
					Http: &mesh_proto.RateLimit_Conf_Http{
						Requests: 100,
						Interval: util_proto.Duration(time.Second * 3),
					},
				},
			},
		}
	}

	type testCase struct {
		dataplane *mesh.DataplaneResource
		policies  []*mesh.RateLimitResource
		expected  core_xds.RateLimitsMap
	}

	DescribeTable("should find best matched policy",
		func(given testCase) {
			manager := core_manager.NewResourceManager(memory.NewStore())
			matcher := RateLimitMatcher{ResourceManager: manager}

			mesh := mesh.NewMeshResource()
			err := manager.Create(context.Background(), mesh, store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			for _, p := range given.policies {
				err := manager.Create(context.Background(), p, store.CreateByKey(p.Meta.GetName(), "default"))
				Expect(err).ToNot(HaveOccurred())
			}

			allMatched, err := matcher.Match(context.Background(), given.dataplane, mesh)
			Expect(err).ToNot(HaveOccurred())
			for key := range allMatched {
				Expect(len(allMatched[key])).To(Equal(len(given.expected[key])))
				for i, matched := range allMatched[key] {
					Expect(matched).To(MatchProto(given.expected[key][i]))
				}
			}
		},
		Entry("1 inbound dataplane, 2 policies", testCase{
			dataplane: dataplaneWithInboundsFunc([]*mesh_proto.Dataplane_Networking_Inbound{
				{
					ServicePort: 8080,
					Tags: map[string]string{
						"kuma.io/service":  "web",
						"version":          "0.1",
						"region":           "eu",
						"kuma.io/protocol": "http",
					},
				},
			}),
			policies: []*mesh.RateLimitResource{
				policyWithDestinationsFunc("rl1", time.Unix(1, 0),
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service":  "*",
								"kuma.io/protocol": "http",
							},
						},
					},
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"region":           "us",
								"kuma.io/protocol": "http",
							},
						},
					}),
				policyWithDestinationsFunc("rl2", time.Unix(1, 0),
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service":  "*",
								"kuma.io/protocol": "http",
							},
						},
					},
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service":  "*",
								"kuma.io/protocol": "http",
							},
						},
					}),
			},
			expected: core_xds.RateLimitsMap{
				mesh_proto.InboundInterface{
					WorkloadIP:   "127.0.0.1",
					WorkloadPort: 8080,
				}: []*mesh_proto.RateLimit{
					policyWithDestinationsFunc("rl2", time.Unix(1, 0),
						[]*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service":  "*",
									"kuma.io/protocol": "http",
								},
							},
						},
						[]*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service":  "*",
									"kuma.io/protocol": "http",
								},
							},
						}).Spec,
				}}}),
		Entry("should apply policy only to the first inbound", testCase{
			dataplane: dataplaneWithInboundsFunc([]*mesh_proto.Dataplane_Networking_Inbound{
				{
					ServicePort: 8080,
					Tags: map[string]string{
						"kuma.io/service":  "web",
						"version":          "0.1",
						"region":           "eu",
						"kuma.io/protocol": "http",
					},
				},
				{
					ServicePort: 8081,
					Tags: map[string]string{
						"kuma.io/service":  "web-api",
						"version":          "0.1.2",
						"region":           "us",
						"kuma.io/protocol": "http",
					},
				},
			}),
			policies: []*mesh.RateLimitResource{
				policyWithDestinationsFunc("rl1", time.Unix(1, 0),
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service":  "*",
								"kuma.io/protocol": "http",
							},
						},
					},
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service":  "web-api",
								"kuma.io/protocol": "http",
							},
						},
					}),
			},
			expected: core_xds.RateLimitsMap{
				mesh_proto.InboundInterface{
					WorkloadIP:   "127.0.0.1",
					WorkloadPort: 8081,
				}: []*mesh_proto.RateLimit{policyWithDestinationsFunc("rl1", time.Unix(1, 0),
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service":  "*",
								"kuma.io/protocol": "http",
							},
						},
					},
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service":  "web-api",
								"kuma.io/protocol": "http",
							},
						},
					}).Spec,
				},
			},
		}),
		Entry("match 2 policies", testCase{
			dataplane: dataplaneWithInboundsFunc([]*mesh_proto.Dataplane_Networking_Inbound{
				{
					ServicePort: 8080,
					Tags: map[string]string{
						"kuma.io/service":  "backend",
						"version":          "0.1",
						"region":           "eu",
						"kuma.io/protocol": "http",
					},
				},
			}),
			policies: []*mesh.RateLimitResource{
				policyWithDestinationsFunc("rl2", time.Unix(1, 0),
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "*",
							},
						},
					},
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service":  "backend",
								"kuma.io/protocol": "http",
							},
						},
					}),
				policyWithDestinationsFunc("rl1", time.Unix(1, 0),
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "frontend",
							},
						},
					},
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service":  "backend",
								"kuma.io/protocol": "http",
							},
						},
					}),
			},
			expected: core_xds.RateLimitsMap{
				mesh_proto.InboundInterface{
					WorkloadIP:   "127.0.0.1",
					WorkloadPort: 8080,
				}: []*mesh_proto.RateLimit{
					policyWithDestinationsFunc("rl2", time.Unix(1, 0),
						[]*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service": "frontend",
								},
							},
						},
						[]*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service":  "backend",
									"kuma.io/protocol": "http",
								},
							},
						}).Spec,
					policyWithDestinationsFunc("rl2", time.Unix(1, 0),
						[]*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service": "*",
								},
							},
						},
						[]*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service":  "backend",
									"kuma.io/protocol": "http",
								},
							},
						}).Spec,
				}}}),
		Entry("match and sort 3 policies", testCase{
			dataplane: dataplaneWithInboundsFunc([]*mesh_proto.Dataplane_Networking_Inbound{
				{
					ServicePort: 8080,
					Tags: map[string]string{
						"kuma.io/service":  "backend",
						"version":          "0.1",
						"region":           "eu",
						"kuma.io/protocol": "http",
					},
				},
			}),
			policies: []*mesh.RateLimitResource{
				policyWithDestinationsFunc("rl1", time.Unix(1, 0),
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "*",
							},
						},
					},
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service":  "backend",
								"kuma.io/protocol": "http",
							},
						},
					}),
				policyWithDestinationsFunc("rl2", time.Unix(1, 0),
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "frontend",
							},
						},
						{
							Match: map[string]string{
								"kuma.io/service": "something_else",
							},
						},
					},
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service":  "backend",
								"kuma.io/protocol": "http",
							},
						},
					}),
				policyWithDestinationsFunc("rl3", time.Unix(1, 0),
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "frontend",
								"kuma.io/zone":    "eu",
							},
						},
					},
					[]*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service":  "backend",
								"kuma.io/protocol": "http",
							},
						},
					}),
			},
			expected: core_xds.RateLimitsMap{
				mesh_proto.InboundInterface{
					WorkloadIP:   "127.0.0.1",
					WorkloadPort: 8080,
				}: []*mesh_proto.RateLimit{
					policyWithDestinationsFunc("rl3", time.Unix(1, 0),
						[]*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service": "frontend",
									"kuma.io/zone":    "eu",
								},
							},
						},
						[]*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service":  "backend",
									"kuma.io/protocol": "http",
								},
							},
						}).Spec,
					policyWithDestinationsFunc("rl2", time.Unix(1, 0),
						[]*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service": "frontend",
								},
							},
						},
						[]*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service":  "backend",
									"kuma.io/protocol": "http",
								},
							},
						}).Spec,
					policyWithDestinationsFunc("rl2", time.Unix(1, 0),
						[]*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service": "something_else",
								},
							},
						},
						[]*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service":  "backend",
									"kuma.io/protocol": "http",
								},
							},
						}).Spec,
					policyWithDestinationsFunc("rl1", time.Unix(1, 0),
						[]*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service": "*",
								},
							},
						},
						[]*mesh_proto.Selector{
							{
								Match: map[string]string{
									"kuma.io/service":  "backend",
									"kuma.io/protocol": "http",
								},
							},
						}).Spec,
				}}}),
	)
})
