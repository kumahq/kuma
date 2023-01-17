package v3_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("FaultInjectionConfigurer", func() {
	type testCase struct {
		input    []*core_mesh.FaultInjectionResource
		expected string
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			filterChain, err := NewFilterChainBuilder(envoy.APIV3).
				Configure(HttpConnectionManager("stats", false)).
				Configure(FaultInjection(given.input...)).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(filterChain)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic input", testCase{
			input: []*core_mesh.FaultInjectionResource{{
				Spec: &mesh_proto.FaultInjection{
					Sources: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"tag1": "value1",
								"tag2": "value2",
							},
						},
					},
					Conf: &mesh_proto.FaultInjection_Conf{
						Delay: &mesh_proto.FaultInjection_Conf_Delay{
							Percentage: util_proto.Double(50),
							Value:      util_proto.Duration(time.Second * 5),
						},
					},
				},
			}},

			expected: `
            filters:
            - name: envoy.filters.network.http_connection_manager
              typedConfig:
                '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                httpFilters:
                - name: envoy.filters.http.fault
                  typedConfig:
                    '@type': type.googleapis.com/envoy.extensions.filters.http.fault.v3.HTTPFault
                    delay:
                      fixedDelay: 5s
                      percentage:
                        numerator: 50
                    headers:
                    - name: x-kuma-tags
                      safeRegexMatch:
                        googleRe2: {}
                        regex: '.*&tag1=[^&]*value1[,&].*&tag2=[^&]*value2[,&].*'
                - name: envoy.filters.http.router
                  typedConfig:
                    '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                statPrefix: stats`,
		}),
		Entry("2 policy selectors", testCase{
			input: []*core_mesh.FaultInjectionResource{{
				Spec: &mesh_proto.FaultInjection{
					Sources: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"tag1": "value1m1",
								"tag2": "value2m1",
							},
						},
						{
							Match: map[string]string{
								"tag1": "value1m2",
								"tag2": "value2m2",
							},
						},
					},
					Conf: &mesh_proto.FaultInjection_Conf{
						Delay: &mesh_proto.FaultInjection_Conf_Delay{
							Percentage: util_proto.Double(50),
							Value:      util_proto.Duration(time.Second * 5),
						},
					},
				},
			}},
			expected: `
            filters:
            - name: envoy.filters.network.http_connection_manager
              typedConfig:
                '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                httpFilters:
                - name: envoy.filters.http.fault
                  typedConfig:
                    '@type': type.googleapis.com/envoy.extensions.filters.http.fault.v3.HTTPFault
                    delay:
                      fixedDelay: 5s
                      percentage:
                        numerator: 50
                    headers:
                    - name: x-kuma-tags
                      safeRegexMatch:
                        googleRe2: {}
                        regex: '(.*&tag1=[^&]*value1m1[,&].*&tag2=[^&]*value2m1[,&].*|.*&tag1=[^&]*value1m2[,&].*&tag2=[^&]*value2m2[,&].*)'
                - name: envoy.filters.http.router
                  typedConfig:
                    '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                statPrefix: stats`,
		}),
		Entry("should preserve the order of passed policies", testCase{
			input: []*core_mesh.FaultInjectionResource{
				{
					Spec: &mesh_proto.FaultInjection{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"tag1": "value1",
									"tag2": "value2",
									"tag3": "value3",
								},
							},
						},
						Conf: &mesh_proto.FaultInjection_Conf{
							Delay: &mesh_proto.FaultInjection_Conf_Delay{
								Percentage: util_proto.Double(100),
								Value:      util_proto.Duration(time.Second * 15),
							},
						},
					},
				},
				{
					Spec: &mesh_proto.FaultInjection{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"tag1": "value1",
									"tag2": "value2",
								},
							},
						},
						Conf: &mesh_proto.FaultInjection_Conf{
							Delay: &mesh_proto.FaultInjection_Conf_Delay{
								Percentage: util_proto.Double(100),
								Value:      util_proto.Duration(time.Second * 10),
							},
						},
					},
				},
				{
					Spec: &mesh_proto.FaultInjection{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"tag1": "*",
									"tag2": "*",
								},
							},
						},
						Conf: &mesh_proto.FaultInjection_Conf{
							Delay: &mesh_proto.FaultInjection_Conf_Delay{
								Percentage: util_proto.Double(100),
								Value:      util_proto.Duration(time.Second * 5),
							},
						},
					},
				},
			},
			expected: `
            filters:
             - name: envoy.filters.network.http_connection_manager
               typedConfig:
                 '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                 httpFilters:
                 - name: envoy.filters.http.fault
                   typedConfig:
                     '@type': type.googleapis.com/envoy.extensions.filters.http.fault.v3.HTTPFault
                     delay:
                       fixedDelay: 15s
                       percentage:
                         numerator: 100
                     headers:
                     - name: x-kuma-tags
                       safeRegexMatch:
                         googleRe2: {}
                         regex: .*&tag1=[^&]*value1[,&].*&tag2=[^&]*value2[,&].*&tag3=[^&]*value3[,&].*
                 - name: envoy.filters.http.fault
                   typedConfig:
                     '@type': type.googleapis.com/envoy.extensions.filters.http.fault.v3.HTTPFault
                     delay:
                       fixedDelay: 10s
                       percentage:
                         numerator: 100
                     headers:
                     - name: x-kuma-tags
                       safeRegexMatch:
                         googleRe2: {}
                         regex: .*&tag1=[^&]*value1[,&].*&tag2=[^&]*value2[,&].*
                 - name: envoy.filters.http.fault
                   typedConfig:
                     '@type': type.googleapis.com/envoy.extensions.filters.http.fault.v3.HTTPFault
                     delay:
                       fixedDelay: 5s
                       percentage:
                         numerator: 100
                     headers:
                     - name: x-kuma-tags
                       safeRegexMatch:
                         googleRe2: {}
                         regex: .*&tag1=.*&tag2=.*
                 - name: envoy.filters.http.router
                   typedConfig:
                     '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                 statPrefix: stats`,
		}),
	)
})
