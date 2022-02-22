package ratelimits_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	. "github.com/kumahq/kuma/pkg/core/ratelimits"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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
	dataplaneWithInboundsFunc := func(inbounds []*mesh_proto.Dataplane_Networking_Inbound) *core_mesh.DataplaneResource {
		return &core_mesh.DataplaneResource{
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

	dataplaneWithOutboundsFunc := func(outbounds []*mesh_proto.Dataplane_Networking_Outbound) *core_mesh.DataplaneResource {
		return &core_mesh.DataplaneResource{
			Meta: &model.ResourceMeta{
				Mesh: "default",
				Name: "dp1",
			},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Outbound: outbounds,
				},
			},
		}
	}

	policyWithDestinationsFunc := func(name string, creationTime time.Time, sources, destinations []*mesh_proto.Selector) *core_mesh.RateLimitResource {
		return &core_mesh.RateLimitResource{
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
		dataplane *core_mesh.DataplaneResource
		policies  []*core_mesh.RateLimitResource
		expected  core_xds.RateLimitsMap
	}

	DescribeTable("should find best matched policy",
		func(given testCase) {
			manager := core_manager.NewResourceManager(memory.NewStore())
			matcher := RateLimitMatcher{ResourceManager: manager}

			mesh := core_mesh.NewMeshResource()
			err := manager.Create(context.Background(), mesh, store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			for _, p := range given.policies {
				err := manager.Create(context.Background(), p, store.CreateByKey(p.Meta.GetName(), "default"))
				Expect(err).ToNot(HaveOccurred())
			}

			allMatched, err := matcher.Match(context.Background(), given.dataplane, mesh)
			Expect(err).ToNot(HaveOccurred())
			inboundMatched := allMatched.Inbound
			for key := range inboundMatched {
				Expect(len(inboundMatched[key])).To(Equal(len(given.expected.Inbound[key])))
				for i, matched := range inboundMatched[key] {
					Expect(matched.Spec).To(MatchProto(given.expected.Inbound[key][i].Spec))
				}
			}
			outboundMatched := allMatched.Outbound
			for i, matched := range outboundMatched {
				Expect(matched).To(MatchProto(given.expected.Outbound[i]))
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
			policies: []*core_mesh.RateLimitResource{
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
				Inbound: core_xds.InboundRateLimitsMap{
					mesh_proto.InboundInterface{
						WorkloadIP:   "127.0.0.1",
						WorkloadPort: 8080,
					}: []*core_mesh.RateLimitResource{
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
					}}}}),

		Entry("1 outbound dataplane, 2 policies", testCase{
			dataplane: dataplaneWithOutboundsFunc([]*mesh_proto.Dataplane_Networking_Outbound{
				{
					Port: 8080,
					Tags: map[string]string{
						"kuma.io/service":  "web",
						"version":          "0.1",
						"region":           "eu",
						"kuma.io/protocol": "http",
					},
				},
			}),
			policies: []*core_mesh.RateLimitResource{
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
				Outbound: core_xds.OutboundRateLimitsMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 8080,
					}: policyWithDestinationsFunc("rl2", time.Unix(1, 0),
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
			},
		}),
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
			policies: []*core_mesh.RateLimitResource{
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
				Inbound: core_xds.InboundRateLimitsMap{
					mesh_proto.InboundInterface{
						WorkloadIP:   "127.0.0.1",
						WorkloadPort: 8081,
					}: []*core_mesh.RateLimitResource{
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
				},
			},
		}),
		Entry("should apply policy only to the first outbound", testCase{
			dataplane: dataplaneWithOutboundsFunc([]*mesh_proto.Dataplane_Networking_Outbound{
				{
					Port: 8080,
					Tags: map[string]string{
						"kuma.io/service":  "web",
						"version":          "0.1",
						"region":           "eu",
						"kuma.io/protocol": "http",
					},
				},
				{
					Port: 8081,
					Tags: map[string]string{
						"kuma.io/service":  "web-api",
						"version":          "0.1.2",
						"region":           "us",
						"kuma.io/protocol": "http",
					},
				},
			}),
			policies: []*core_mesh.RateLimitResource{
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
				Outbound: core_xds.OutboundRateLimitsMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 8081,
					}: policyWithDestinationsFunc("rl1", time.Unix(1, 0),
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
			},
		}),

		Entry("match 2 policies on inbound", testCase{
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
			policies: []*core_mesh.RateLimitResource{
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
				Inbound: core_xds.InboundRateLimitsMap{
					mesh_proto.InboundInterface{
						WorkloadIP:   "127.0.0.1",
						WorkloadPort: 8080,
					}: []*core_mesh.RateLimitResource{
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
							}),
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
					}}}}),

		Entry("match 2 policies on outbound", testCase{
			dataplane: dataplaneWithOutboundsFunc([]*mesh_proto.Dataplane_Networking_Outbound{
				{
					Port: 8080,
					Tags: map[string]string{
						"kuma.io/service":  "backend",
						"version":          "0.1",
						"region":           "eu",
						"kuma.io/protocol": "http",
					},
				},
			}),
			policies: []*core_mesh.RateLimitResource{
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
				Outbound: core_xds.OutboundRateLimitsMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 8080,
					}: policyWithDestinationsFunc("rl2", time.Unix(1, 0),
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
			},
		}),

		Entry("match and sort 3 policies inbound", testCase{
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
			policies: []*core_mesh.RateLimitResource{
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
				Inbound: core_xds.InboundRateLimitsMap{
					mesh_proto.InboundInterface{
						WorkloadIP:   "127.0.0.1",
						WorkloadPort: 8080,
					}: []*core_mesh.RateLimitResource{
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
							}),
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
							}),
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
					}}}}),

		Entry("match and sort 3 policies outbound", testCase{
			dataplane: dataplaneWithOutboundsFunc([]*mesh_proto.Dataplane_Networking_Outbound{
				{
					Port: 8080,
					Tags: map[string]string{
						"kuma.io/service":  "backend",
						"version":          "0.1",
						"region":           "eu",
						"kuma.io/protocol": "http",
					},
				},
			}),
			policies: []*core_mesh.RateLimitResource{
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
				Outbound: core_xds.OutboundRateLimitsMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 8080,
					}: policyWithDestinationsFunc("rl3", time.Unix(1, 0),
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
			},
		},
		),
	)
})
