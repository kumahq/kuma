package listeners_test

import (
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	. "github.com/Kong/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("FaultInjectionConfigurer", func() {
	type testCase struct {
		input    *mesh_proto.FaultInjection
		expected string
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			filterChain, err := NewFilterChainBuilder().
				Configure(HttpConnectionManager("stats")).
				Configure(FaultInjection(given.input)).
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
			input: &mesh_proto.FaultInjection{
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
						Percentage: &wrappers.DoubleValue{Value: 50},
						Value:      &duration.Duration{Seconds: 5},
					},
				},
			},

			expected: `
            filters:
            - name: envoy.http_connection_manager
              typedConfig:
                '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
                httpFilters:
                - name: envoy.fault
                  typedConfig:
                    '@type': type.googleapis.com/envoy.config.filter.http.fault.v2.HTTPFault
                    delay:
                      fixedDelay: 5s
                      percentage:
                        numerator: 50
                    headers:
                    - name: x-kuma-tags
                      safeRegexMatch:
                        googleRe2: 
                          maxProgramSize: 500
                        regex: '.*&tag1=[^&]*value1[,&].*&tag2=[^&]*value2[,&].*'
                - name: envoy.router
                statPrefix: stats`,
		}),
		Entry("2 policy selectors", testCase{
			input: &mesh_proto.FaultInjection{
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
						Percentage: &wrappers.DoubleValue{Value: 50},
						Value:      &duration.Duration{Seconds: 5},
					},
				},
			},

			expected: `
            filters:
            - name: envoy.http_connection_manager
              typedConfig:
                '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
                httpFilters:
                - name: envoy.fault
                  typedConfig:
                    '@type': type.googleapis.com/envoy.config.filter.http.fault.v2.HTTPFault
                    delay:
                      fixedDelay: 5s
                      percentage:
                        numerator: 50
                    headers:
                    - name: x-kuma-tags
                      safeRegexMatch:
                        googleRe2: 
                          maxProgramSize: 500
                        regex: '(.*&tag1=[^&]*value1m1[,&].*&tag2=[^&]*value2m1[,&].*|.*&tag1=[^&]*value1m2[,&].*&tag2=[^&]*value2m2[,&].*)'
                - name: envoy.router
                statPrefix: stats`,
		}),
	)
})
