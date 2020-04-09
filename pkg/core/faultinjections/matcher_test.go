package faultinjections_test

import (
	"context"
	"time"

	core_xds "github.com/Kong/kuma/pkg/core/xds"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	. "github.com/Kong/kuma/pkg/core/faultinjections"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("Match", func() {
	dataplaneWithInboundsFunc := func(inbounds []*mesh_proto.Dataplane_Networking_Inbound) *mesh.DataplaneResource {
		return &mesh.DataplaneResource{
			Meta: &model.ResourceMeta{
				Mesh: "default",
				Name: "dp1",
			},
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Inbound: inbounds,
				},
			},
		}
	}

	policyWithDestinationsFunc := func(name string, creationTime time.Time, destinations []*mesh_proto.Selector) *mesh.FaultInjectionResource {
		return &mesh.FaultInjectionResource{
			Meta: &model.ResourceMeta{
				Name:         name,
				CreationTime: creationTime,
			},
			Spec: mesh_proto.FaultInjection{
				Sources: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service":  "*",
							"protocol": "http",
						},
					},
				},
				Destinations: destinations,
				Conf: &mesh_proto.FaultInjection_Conf{
					Delay: &mesh_proto.FaultInjection_Conf_Delay{
						Percentage: &wrappers.DoubleValue{Value: 50},
						Value:      &duration.Duration{Seconds: 5},
					},
				},
			},
		}
	}

	type testCase struct {
		dataplane *mesh.DataplaneResource
		policies  []*mesh.FaultInjectionResource
		expected  core_xds.FaultInjectionMap
	}

	DescribeTable("should find best matched policy",
		func(given testCase) {
			manager := core_manager.NewResourceManager(memory.NewStore())
			matcher := FaultInjectionMatcher{ResourceManager: manager}

			err := manager.Create(context.Background(), &mesh.MeshResource{}, store.CreateByKey("default", "default"))
			Expect(err).ToNot(HaveOccurred())

			for _, p := range given.policies {
				err := manager.Create(context.Background(), p, store.CreateByKey(p.Meta.GetName(), "default"))
				Expect(err).ToNot(HaveOccurred())
			}

			bestMatched, err := matcher.Match(context.Background(), given.dataplane)
			Expect(err).ToNot(HaveOccurred())
			Expect(bestMatched).To(Equal(given.expected))
		},
		Entry("1 inbound dataplane, 2 policies", testCase{
			dataplane: dataplaneWithInboundsFunc([]*mesh_proto.Dataplane_Networking_Inbound{
				{
					ServicePort: 8080,
					Tags: map[string]string{
						"service":  "web",
						"version":  "0.1",
						"region":   "eu",
						"protocol": "http",
					},
				},
			}),
			policies: []*mesh.FaultInjectionResource{
				policyWithDestinationsFunc("fi1", time.Unix(1, 0), []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"region":   "us",
							"protocol": "http",
						},
					},
				}),
				policyWithDestinationsFunc("fi2", time.Unix(1, 0), []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service":  "*",
							"protocol": "http",
						},
					},
				}),
			},
			expected: core_xds.FaultInjectionMap{
				mesh_proto.InboundInterface{
					WorkloadPort: 8080,
				}: &policyWithDestinationsFunc("fi2", time.Unix(1, 0), []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service":  "*",
							"protocol": "http",
						},
					},
				}).Spec,
			}}),
		Entry("should apply policy only to the first inbound", testCase{
			dataplane: dataplaneWithInboundsFunc([]*mesh_proto.Dataplane_Networking_Inbound{
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
			}),
			policies: []*mesh.FaultInjectionResource{
				policyWithDestinationsFunc("fi1", time.Unix(1, 0), []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service":  "web-api",
							"protocol": "http",
						},
					},
				}),
			},
			expected: core_xds.FaultInjectionMap{
				mesh_proto.InboundInterface{
					WorkloadPort: 8081,
				}: &policyWithDestinationsFunc("fi1", time.Unix(1, 0), []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service":  "web-api",
							"protocol": "http",
						},
					},
				}).Spec,
			},
		}),
	)
})
