package listeners_test

import (
	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	. "github.com/Kong/kuma/pkg/xds/envoy/listeners"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("FaultInjectionConfigurer", func() {
	type testCase struct {
		input    *mesh.FaultInjectionResource
		expected string
	}
	DescribeTable("",
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
			input: &mesh.FaultInjectionResource{
				Spec: v1alpha1.FaultInjection{
					Conf: &v1alpha1.FaultInjection_Conf{
						Delay: &v1alpha1.FaultInjection_Conf_Delay{
							Percentage: &wrappers.DoubleValue{Value: 50},
							Value:      &duration.Duration{Seconds: 5},
						},
					},
				},
			},
			expected: `
            filters:
            - name: envoy.http_connection_manager
              typedConfig:
                '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
                httpFilters:
                - name: envoy.router
                - name: envoy.filters.http.fault
                  typedConfig:
                    '@type': type.googleapis.com/envoy.config.filter.http.fault.v2.HTTPFault
                    delay:
                      fixedDelay: 5s
                      percentage:
                        numerator: 50
                statPrefix: stats`,
		}),
	)
})
