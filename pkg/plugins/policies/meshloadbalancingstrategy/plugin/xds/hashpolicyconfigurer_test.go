package xds_test

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/plugin/xds"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

var _ = Describe("HashPolicyConfigurer", func() {
	type testCase struct {
		hashPolicies []api.HashPolicy
		expected     string
	}

	DescribeTable("should generate proper envoy config",
		func(given testCase) {
			// given
			configurer := &xds.HashPolicyConfigurer{
				HashPolicies: given.hashPolicies,
			}
			rb, err := envoy_routes.NewRouteBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
				Configure(
					envoy_routes.RouteMustConfigureFunc(func(envoyRoute *envoy_route.Route) {
						envoyRoute.Action = &envoy_route.Route_Route{
							Route: &envoy_route.RouteAction{},
						}
					})).Build()
			Expect(err).ToNot(HaveOccurred())

			// when
			err = configurer.Configure(rb.(*envoy_route.Route))
			Expect(err).ToNot(HaveOccurred())

			// then
			actual, err := util_proto.ToYAML(rb)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("header hash policy", testCase{
			hashPolicies: []api.HashPolicy{
				{
					Type: api.HeaderType,
					Header: &api.Header{
						Name: "x-header",
					},
				},
			},
			expected: `
match: {}
route:
  hashPolicy:
    - header:
        headerName: x-header`,
		}),
		Entry("cookie hash policy", testCase{
			hashPolicies: []api.HashPolicy{
				{
					Type: api.CookieType,
					Cookie: &api.Cookie{
						Name: "x-header",
						TTL:  test.ParseDuration("1h"),
						Path: pointer.To("/api"),
					},
				},
			},
			expected: `
match: {}
route:
  hashPolicy:
    - cookie:
        name: x-header
        path: /api
        ttl: 3600s`,
		}),
		Entry("connection hash policy", testCase{
			hashPolicies: []api.HashPolicy{
				{
					Type: api.ConnectionType,
					Connection: &api.Connection{
						SourceIP: pointer.To(true),
					},
				},
			},
			expected: `
match: {}
route:
  hashPolicy:
    - connectionProperties:
        sourceIp: true`,
		}),
		Entry("query hash policy", testCase{
			hashPolicies: []api.HashPolicy{
				{
					Type: api.QueryParameterType,
					QueryParameter: &api.QueryParameter{
						Name: "queryparam",
					},
				},
			},
			expected: `
match: {}
route:
  hashPolicy:
    - queryParameter:
        name: queryparam`,
		}),
		Entry("filter state hash policy", testCase{
			hashPolicies: []api.HashPolicy{
				{
					Type: api.FilterStateType,
					FilterState: &api.FilterState{
						Key: "filterstate-key",
					},
				},
			},
			expected: `
match: {}
route:
  hashPolicy:
    - filterState:
        key: filterstate-key`,
		}),
		Entry("multiple hash policies", testCase{
			hashPolicies: []api.HashPolicy{
				{
					Type: api.FilterStateType,
					FilterState: &api.FilterState{
						Key: "filterstate-key",
					},
					Terminal: pointer.To(true),
				},
				{
					Type: api.QueryParameterType,
					QueryParameter: &api.QueryParameter{
						Name: "queryparam",
					},
					Terminal: pointer.To(false),
				},
				{
					Type: api.ConnectionType,
					Connection: &api.Connection{
						SourceIP: pointer.To(true),
					},
					Terminal: pointer.To(true),
				},
				{
					Type: api.CookieType,
					Cookie: &api.Cookie{
						Name: "x-header",
						TTL:  test.ParseDuration("1h"),
						Path: pointer.To("/api"),
					},
				},
				{
					Type: api.HeaderType,
					Header: &api.Header{
						Name: "x-header",
					},
					Terminal: pointer.To(true),
				},
			},
			expected: `
match: {}
route:
  hashPolicy:
    - filterState:
        key: filterstate-key
      terminal: true
    - queryParameter:
        name: queryparam
    - connectionProperties:
        sourceIp: true
      terminal: true
    - cookie:
        name: x-header
        path: /api
        ttl: 3600s
    - header:
        headerName: x-header
      terminal: true`,
		}),
	)
})
