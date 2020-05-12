package xds_test

import (
	"fmt"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

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
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {},
						envoy_types.Runtime: envoy_cache.NewResources("", []envoy_types.Resource{}),
					},
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("101", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("102", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("103", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("104", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {Version: "105"},
						envoy_types.Runtime: envoy_cache.NewResources("106", []envoy_types.Resource{}),
					},
				}},
			}),
			Entry("when 'old' = `nil` and each resource type in 'new' has the same version", testCase{
				old: nil,
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {}, // empty version must be replaced
						envoy_types.Runtime: envoy_cache.NewResources("v1", []envoy_types.Resource{}),
					},
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {Version: "101"},
						envoy_types.Runtime: envoy_cache.NewResources("v1", []envoy_types.Resource{}),
					},
				}},
			}),
			Entry("when 'old' = `nil` and each resource type in 'new' has different version", testCase{
				old: nil,
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("v2", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("v3", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("v4", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {}, // empty version must be replaced
						envoy_types.Runtime: envoy_cache.NewResources("v6", []envoy_types.Resource{}),
					},
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("v2", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("v3", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("v4", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {Version: "101"},
						envoy_types.Runtime: envoy_cache.NewResources("v6", []envoy_types.Resource{}),
					},
				}},
			}),
			Entry("when 'old' != `nil`, resources hasn't changed, versions are empty", testCase{
				old: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("v2", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("v3", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("v4", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {Version: "v5"},
						envoy_types.Runtime: envoy_cache.NewResources("v6", []envoy_types.Resource{}),
					},
				}},
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {Version: ""},
						envoy_types.Runtime: envoy_cache.NewResources("", []envoy_types.Resource{}),
					},
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("v2", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("v3", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("v4", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {Version: "v5"},
						envoy_types.Runtime: envoy_cache.NewResources("v6", []envoy_types.Resource{}),
					},
				}},
			}),
			Entry("when 'old' != `nil`, resources hasn't changed, versions are not empty", testCase{
				old: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("v2", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("v3", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("v4", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {Version: "v5"},
						envoy_types.Runtime: envoy_cache.NewResources("v6", []envoy_types.Resource{}),
					},
				}},
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("v11", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("v22", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("v33", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("v44", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {Version: "v55"},
						envoy_types.Runtime: envoy_cache.NewResources("v66", []envoy_types.Resource{}),
					},
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("v11", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("v22", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("v33", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("v44", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {Version: "v55"},
						envoy_types.Runtime: envoy_cache.NewResources("v66", []envoy_types.Resource{}),
					},
				}},
			}),
			Entry("when 'old' != `nil`, resources deleted, versions are empty", testCase{
				old: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("v2", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("v3", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("v4", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {Version: "v5"},                                          // version should stay the same
						envoy_types.Runtime: envoy_cache.NewResources("v6", []envoy_types.Resource{}), // version should stay the same
					},
				}},
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("", []envoy_types.Resource{}),
						envoy_types.Cluster:  envoy_cache.NewResources("", []envoy_types.Resource{}),
						envoy_types.Route:    envoy_cache.NewResources("", []envoy_types.Resource{}),
						envoy_types.Listener: {Version: ""},
						envoy_types.Secret:   {Version: ""},
						envoy_types.Runtime:  {Version: ""},
					},
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("101", []envoy_types.Resource{}),
						envoy_types.Cluster:  envoy_cache.NewResources("102", []envoy_types.Resource{}),
						envoy_types.Route:    envoy_cache.NewResources("103", []envoy_types.Resource{}),
						envoy_types.Listener: {Version: "104"},
						envoy_types.Secret:   {Version: "v5"},
						envoy_types.Runtime:  {Version: "v6"},
					},
				}},
			}),
			Entry("when 'old' != `nil`, resources added, versions are empty", testCase{
				old: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("v2", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("v3", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("v4", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {Version: "v5"},                                          // version should stay the same
						envoy_types.Runtime: envoy_cache.NewResources("v6", []envoy_types.Resource{}), // version should stay the same
					},
				}},
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment2"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
							&envoy.Cluster{Name: "Cluster2"},
						}),
						envoy_types.Route: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
							&envoy.RouteConfiguration{Name: "RouteConfiguration2"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
							&envoy.Listener{Name: "Listener2"},
						}),
						envoy_types.Secret:  {Version: ""}, // version should stay the same
						envoy_types.Runtime: {Version: ""}, // version should stay the same
					},
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("101", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment2"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("102", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
							&envoy.Cluster{Name: "Cluster2"},
						}),
						envoy_types.Route: envoy_cache.NewResources("103", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
							&envoy.RouteConfiguration{Name: "RouteConfiguration2"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("104", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
							&envoy.Listener{Name: "Listener2"},
						}),
						envoy_types.Secret:  {Version: "v5"},
						envoy_types.Runtime: {Version: "v6"},
					},
				}},
			}),
			Entry("when 'old' != `nil`, resources modified, versions are empty", testCase{
				old: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("v1", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment"},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("v2", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster"},
						}),
						envoy_types.Route: envoy_cache.NewResources("v3", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration"},
						}),
						envoy_types.Listener: envoy_cache.NewResources("v4", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener"},
						}),
						envoy_types.Secret:  {Version: "v5"},                                          // version should stay the same
						envoy_types.Runtime: envoy_cache.NewResources("v6", []envoy_types.Resource{}), // version should stay the same
					},
				}},
				new: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment", Policy: &envoy.ClusterLoadAssignment_Policy{DisableOverprovisioning: true}},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster", AltStatName: "AltStatName"},
						}),
						envoy_types.Route: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration", MostSpecificHeaderMutationsWins: true},
						}),
						envoy_types.Listener: envoy_cache.NewResources("", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener", ContinueOnListenerFiltersTimeout: true},
						}),
						envoy_types.Secret:  {Version: ""}, // version should stay the same
						envoy_types.Runtime: {Version: ""}, // version should stay the same
					},
				}},
				expected: &SampleSnapshot{envoy_cache.Snapshot{
					Resources: [envoy_types.UnknownType]envoy_cache.Resources{
						envoy_types.Endpoint: envoy_cache.NewResources("101", []envoy_types.Resource{
							&envoy.ClusterLoadAssignment{ClusterName: "ClusterLoadAssignment", Policy: &envoy.ClusterLoadAssignment_Policy{DisableOverprovisioning: true}},
						}),
						envoy_types.Cluster: envoy_cache.NewResources("102", []envoy_types.Resource{
							&envoy.Cluster{Name: "Cluster", AltStatName: "AltStatName"},
						}),
						envoy_types.Route: envoy_cache.NewResources("103", []envoy_types.Resource{
							&envoy.RouteConfiguration{Name: "RouteConfiguration", MostSpecificHeaderMutationsWins: true},
						}),
						envoy_types.Listener: envoy_cache.NewResources("104", []envoy_types.Resource{
							&envoy.Listener{Name: "Listener", ContinueOnListenerFiltersTimeout: true},
						}),
						envoy_types.Secret:  {Version: "v5"},
						envoy_types.Runtime: {Version: "v6"},
					},
				}},
			}),
		)
	})
})
