package xds_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"

	. "github.com/Kong/kuma/pkg/util/xds"
)

var _ = Describe("SnapshotAutoVersioner", func() {

	Describe("Version()", func() {

		It("should supoprt `nil`", func() {
			// setup
			versioner := SnapshotAutoVersioner{}

			// when
			actual := versioner.Version(nil, nil)
			// then
			Expect(actual).To(BeNil())

			// when
			actual = versioner.Version(nil, NewSampleSnapshot("v1", nil, nil, nil, nil, nil))
			// then
			Expect(actual).To(BeNil())
		})

		type testCase struct {
			old      Snapshot
			new      Snapshot
			expected Snapshot
		}

		DescribeTable("should infer version of a new Snapshot by comparing against the old one",
			func(given testCase) {
				// setup
				uuid := uint64(101)
				versioner := SnapshotAutoVersioner{
					UUID: func() string {
						defer func() { uuid++ }()
						return fmt.Sprintf("%d", uuid)
					},
				}
				// when
				actual := versioner.Version(given.new, given.old)
				// then
				Expect(actual).To(Equal(given.expected))
			},
			Entry("when 'old' = `nil` and 'new' has empty version", testCase{
				old: nil,
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{},
					Runtimes: envoy_cache.NewResources("", []envoy_cache.Resource{}),
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("101", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("102", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("103", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("104", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{Version: "105"},
					Runtimes: envoy_cache.NewResources("106", []envoy_cache.Resource{}),
				}},
			}),
			Entry("when 'old' = `nil` and each resource type in 'new' has the same version", testCase{
				old: nil,
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{}, // empty version must be replaced
					Runtimes: envoy_cache.NewResources("v1", []envoy_cache.Resource{}),
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{Version: "101"},
					Runtimes: envoy_cache.NewResources("v1", []envoy_cache.Resource{}),
				}},
			}),
			Entry("when 'old' = `nil` and each resource type in 'new' has different version", testCase{
				old: nil,
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("v2", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("v3", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("v4", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{}, // empty version must be replaced
					Runtimes: envoy_cache.NewResources("v6", []envoy_cache.Resource{}),
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("v2", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("v3", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("v4", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{Version: "101"},
					Runtimes: envoy_cache.NewResources("v6", []envoy_cache.Resource{}),
				}},
			}),
			Entry("when 'old' != `nil`, resources hasn't changed, versions are empty", testCase{
				old: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("v2", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("v3", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("v4", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{Version: "v5"},
					Runtimes: envoy_cache.NewResources("v6", []envoy_cache.Resource{}),
				}},
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{Version: ""},
					Runtimes: envoy_cache.NewResources("", []envoy_cache.Resource{}),
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("v2", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("v3", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("v4", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{Version: "v5"},
					Runtimes: envoy_cache.NewResources("v6", []envoy_cache.Resource{}),
				}},
			}),
			Entry("when 'old' != `nil`, resources hasn't changed, versions are not empty", testCase{
				old: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("v2", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("v3", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("v4", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{Version: "v5"},
					Runtimes: envoy_cache.NewResources("v6", []envoy_cache.Resource{}),
				}},
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("v11", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("v22", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("v33", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("v44", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{Version: "v55"},
					Runtimes: envoy_cache.NewResources("v66", []envoy_cache.Resource{}),
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("v11", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("v22", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("v33", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("v44", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{Version: "v55"},
					Runtimes: envoy_cache.NewResources("v66", []envoy_cache.Resource{}),
				}},
			}),
			Entry("when 'old' != `nil`, resources deleted, versions are empty", testCase{
				old: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("v2", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("v3", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("v4", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{Version: "v5"},                     // version should stay the same
					Runtimes: envoy_cache.NewResources("v6", []envoy_cache.Resource{}), // version should stay the same
				}},
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("", []envoy_cache.Resource{}),
					Clusters:  envoy_cache.NewResources("", []envoy_cache.Resource{}),
					Routes:    envoy_cache.NewResources("", []envoy_cache.Resource{}),
					Listeners: envoy_cache.Resources{Version: ""},
					Secrets:   envoy_cache.Resources{Version: ""},
					Runtimes:  envoy_cache.Resources{Version: ""},
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("101", []envoy_cache.Resource{}),
					Clusters:  envoy_cache.NewResources("102", []envoy_cache.Resource{}),
					Routes:    envoy_cache.NewResources("103", []envoy_cache.Resource{}),
					Listeners: envoy_cache.Resources{Version: "104"},
					Secrets:   envoy_cache.Resources{Version: "v5"},
					Runtimes:  envoy_cache.Resources{Version: "v6"},
				}},
			}),
			Entry("when 'old' != `nil`, resources added, versions are empty", testCase{
				old: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("v2", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("v3", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("v4", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{Version: "v5"},                     // version should stay the same
					Runtimes: envoy_cache.NewResources("v6", []envoy_cache.Resource{}), // version should stay the same
				}},
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment2"},
					}),
					Clusters: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
						&envoy.Cluster{Name: "Cluster2"},
					}),
					Routes: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						&envoy.RouteConfiguration{Name: "RouteConfiguration2"},
					}),
					Listeners: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
						&envoy.Listener{Name: "Listener2"},
					}),
					Secrets:  envoy_cache.Resources{Version: ""}, // version should stay the same
					Runtimes: envoy_cache.Resources{Version: ""}, // version should stay the same
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("101", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment2"},
					}),
					Clusters: envoy_cache.NewResources("102", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
						&envoy.Cluster{Name: "Cluster2"},
					}),
					Routes: envoy_cache.NewResources("103", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						&envoy.RouteConfiguration{Name: "RouteConfiguration2"},
					}),
					Listeners: envoy_cache.NewResources("104", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
						&envoy.Listener{Name: "Listener2"},
					}),
					Secrets:  envoy_cache.Resources{Version: "v5"},
					Runtimes: envoy_cache.Resources{Version: "v6"},
				}},
			}),
			Entry("when 'old' != `nil`, resources modified, versions are empty", testCase{
				old: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("v1", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
					}),
					Clusters: envoy_cache.NewResources("v2", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster"},
					}),
					Routes: envoy_cache.NewResources("v3", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration"},
					}),
					Listeners: envoy_cache.NewResources("v4", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener"},
					}),
					Secrets:  envoy_cache.Resources{Version: "v5"},                     // version should stay the same
					Runtimes: envoy_cache.NewResources("v6", []envoy_cache.Resource{}), // version should stay the same
				}},
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment", Policy: &envoy.ClusterLoadAssignment_Policy{DisableOverprovisioning: true}},
					}),
					Clusters: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster", AltStatName: "AltStatName"},
					}),
					Routes: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration", MostSpecificHeaderMutationsWins: true},
					}),
					Listeners: envoy_cache.NewResources("", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener", ContinueOnListenerFiltersTimeout: true},
					}),
					Secrets:  envoy_cache.Resources{Version: ""}, // version should stay the same
					Runtimes: envoy_cache.Resources{Version: ""}, // version should stay the same
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Endpoints: envoy_cache.NewResources("101", []envoy_cache.Resource{
						&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment", Policy: &envoy.ClusterLoadAssignment_Policy{DisableOverprovisioning: true}},
					}),
					Clusters: envoy_cache.NewResources("102", []envoy_cache.Resource{
						&envoy.Cluster{Name: "Cluster", AltStatName: "AltStatName"},
					}),
					Routes: envoy_cache.NewResources("103", []envoy_cache.Resource{
						&envoy.RouteConfiguration{Name: "RouteConfiguration", MostSpecificHeaderMutationsWins: true},
					}),
					Listeners: envoy_cache.NewResources("104", []envoy_cache.Resource{
						&envoy.Listener{Name: "Listener", ContinueOnListenerFiltersTimeout: true},
					}),
					Secrets:  envoy_cache.Resources{Version: "v5"},
					Runtimes: envoy_cache.Resources{Version: "v6"},
				}},
			}),
		)
	})
})
