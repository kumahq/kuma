package faultinjections_test

import (
	"context"
	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/faultinjections"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/test/resources/model"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Match", func() {
	dataplaneWithTagsFunc := func(tags map[string]string) *mesh.DataplaneResource {
		return &mesh.DataplaneResource{
			Meta: &model.ResourceMeta{
				Mesh: "default",
				Name: "dp1",
			},
			Spec: v1alpha1.Dataplane{
				Networking: &v1alpha1.Dataplane_Networking{
					Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
						{
							Tags: tags,
						},
					},
				},
			},
		}
	}

	policyWithDestinationsFunc := func(name string, creationTime time.Time, destinations map[string]string) *mesh.FaultInjectionResource {
		return &mesh.FaultInjectionResource{
			Meta: &model.ResourceMeta{
				Name:         name,
				CreationTime: creationTime,
			},
			Spec: v1alpha1.FaultInjection{
				Sources: []*v1alpha1.Selector{
					{
						Match: map[string]string{
							"service":  "*",
							"protocol": "http",
						},
					},
				},
				Destinations: []*v1alpha1.Selector{
					{
						Match: destinations,
					},
				},
				Conf: &v1alpha1.FaultInjection_Conf{
					Delay: &v1alpha1.FaultInjection_Conf_Delay{
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
		expected  string
	}

	DescribeTable("should find best matched policy",
		func(given testCase) {
			manager := core_manager.NewResourceManager(memory.NewStore())
			matcher := faultinjections.FaultInjectionMatcher{ResourceManager: manager}

			err := manager.Create(context.Background(), &mesh.MeshResource{}, store.CreateByKey("default", "default"))
			Expect(err).ToNot(HaveOccurred())

			for _, p := range given.policies {
				err := manager.Create(context.Background(), p, store.CreateByKey(p.Meta.GetName(), "default"))
				Expect(err).ToNot(HaveOccurred())
			}

			bestMatched, err := matcher.Match(context.Background(), given.dataplane)
			Expect(err).ToNot(HaveOccurred())
			Expect(bestMatched.GetMeta().GetName()).To(Equal(given.expected))
		},
		Entry("basic", testCase{
			dataplane: dataplaneWithTagsFunc(map[string]string{
				"service":  "web",
				"version":  "0.1",
				"region":   "eu",
				"protocol": "http",
			}),
			policies: []*mesh.FaultInjectionResource{
				policyWithDestinationsFunc("fi1", time.Unix(1, 0), map[string]string{
					"region":   "us",
					"protocol": "http",
				}),
				policyWithDestinationsFunc("fi2", time.Unix(1, 0), map[string]string{
					"service":  "*",
					"protocol": "http",
				}),
			},
			expected: "fi2"}),
	)
})
