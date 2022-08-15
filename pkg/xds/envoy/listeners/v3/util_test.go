package v3_test

import (
	"errors"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

var _ = Describe("UpdateFilterConfig()", func() {

	Context("happy path", func() {
		type testCase struct {
			filterChain *envoy_listener.FilterChain
			filterName  string
			updateFunc  func(proto.Message) error
			expected    string
		}

		DescribeTable("should update filter config",
			func(given testCase) {
				// when
				err := UpdateFilterConfig(given.filterChain, given.filterName, given.updateFunc)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				actual, err := util_proto.ToYAML(given.filterChain)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("0 filters", testCase{
				filterChain: &envoy_listener.FilterChain{},
				filterName:  "envoy.filters.network.tcp_proxy",
				updateFunc:  func(proto.Message) error { return errors.New("should never happen") },
				expected:    `{}`,
			}),
			Entry("1 filter", func() testCase {
				pbst, err := anypb.New(&envoy_tcp.TcpProxy{})
				Expect(err).ToNot(HaveOccurred())
				return testCase{
					filterChain: &envoy_listener.FilterChain{
						Filters: []*envoy_listener.Filter{{
							Name: "envoy.filters.network.tcp_proxy",
							ConfigType: &envoy_listener.Filter_TypedConfig{
								TypedConfig: pbst,
							},
						}},
					},
					filterName: "envoy.filters.network.tcp_proxy",
					updateFunc: func(filterConfig proto.Message) error {
						proxy := filterConfig.(*envoy_tcp.TcpProxy)
						proxy.ClusterSpecifier = &envoy_tcp.TcpProxy_Cluster{
							Cluster: "backend",
						}
						return nil
					},
					expected: `
                    filters:
                    - name: envoy.filters.network.tcp_proxy
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                        cluster: backend
`,
				}
			}()),
			Entry("2 filters", func() testCase {
				pbst, err := anypb.New(&envoy_tcp.TcpProxy{})
				Expect(err).ToNot(HaveOccurred())
				return testCase{
					filterChain: &envoy_listener.FilterChain{
						Filters: []*envoy_listener.Filter{
							{
								Name: "envoy.filters.network.rbac",
							},
							{
								Name: "envoy.filters.network.tcp_proxy",
								ConfigType: &envoy_listener.Filter_TypedConfig{
									TypedConfig: pbst,
								},
							},
						},
					},
					filterName: "envoy.filters.network.tcp_proxy",
					updateFunc: func(filterConfig proto.Message) error {
						proxy := filterConfig.(*envoy_tcp.TcpProxy)
						proxy.ClusterSpecifier = &envoy_tcp.TcpProxy_Cluster{
							Cluster: "backend",
						}
						return nil
					},
					expected: `
                    filters:
                    - name: envoy.filters.network.rbac
                    - name: envoy.filters.network.tcp_proxy
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                        cluster: backend
`,
				}
			}()),
		)
	})

	Context("error path", func() {

		type testCase struct {
			filterChain *envoy_listener.FilterChain
			filterName  string
			updateFunc  func(proto.Message) error
			expectedErr string
		}

		DescribeTable("should return an error",
			func(given testCase) {
				// when
				err := UpdateFilterConfig(given.filterChain, given.filterName, given.updateFunc)
				// then
				Expect(err).To(HaveOccurred())
				// and
				Expect(err.Error()).To(Equal(given.expectedErr))
			},
			Entry("1 filter without config", testCase{
				filterChain: &envoy_listener.FilterChain{
					Filters: []*envoy_listener.Filter{{
						Name: "envoy.filters.network.tcp_proxy",
					}},
				},
				filterName:  "envoy.filters.network.tcp_proxy",
				updateFunc:  func(proto.Message) error { return errors.New("should never happen") },
				expectedErr: `filters[0]: config cannot be 'nil'`,
			}),
			Entry("1 filter with a wrong config type", func() testCase {
				pbst, err := anypb.New(&envoy_hcm.HttpConnectionManager{})
				Expect(err).ToNot(HaveOccurred())
				return testCase{
					filterChain: &envoy_listener.FilterChain{
						Filters: []*envoy_listener.Filter{{
							Name: "envoy.filters.network.tcp_proxy",
							ConfigType: &envoy_listener.Filter_TypedConfig{
								TypedConfig: pbst,
							},
						}},
					},
					filterName:  "envoy.filters.network.tcp_proxy",
					updateFunc:  func(proto.Message) error { return errors.New("wrong config type") },
					expectedErr: `wrong config type`,
				}
			}()),
		)
	})
})

var _ = Describe("NewUnexpectedFilterConfigTypeError()", func() {

	type testCase struct {
		inputActual   proto.Message
		inputExpected proto.Message
		expectedErr   string
	}

	DescribeTable("should generate proper error message",
		func(given testCase) {
			// when
			err := NewUnexpectedFilterConfigTypeError(given.inputActual, given.inputExpected)
			// then
			Expect(err).To(HaveOccurred())
			// and
			Expect(err.Error()).To(ContainSubstring(given.expectedErr))
		},
		Entry("TcpProxy instead of HttpConnectionManager", testCase{
			inputActual:   &envoy_tcp.TcpProxy{},
			inputExpected: &envoy_hcm.HttpConnectionManager{},
			expectedErr:   `filter config has unexpected type`,
		}),
	)
})

var _ = Describe("ConvertPercentage", func() {
	type testCase struct {
		input    *wrapperspb.DoubleValue
		expected *envoy_type.FractionalPercent
	}
	DescribeTable("should properly converts from percent to fractional percen",
		func(given testCase) {
			fpercent := ConvertPercentage(given.input)
			Expect(fpercent).To(Equal(given.expected))
		},
		Entry("integer input", testCase{
			input:    util_proto.Double(50),
			expected: &envoy_type.FractionalPercent{Numerator: 50, Denominator: envoy_type.FractionalPercent_HUNDRED},
		}),
		Entry("fractional input with 1 digit after dot", testCase{
			input:    util_proto.Double(50.1),
			expected: &envoy_type.FractionalPercent{Numerator: 501000, Denominator: envoy_type.FractionalPercent_TEN_THOUSAND},
		}),
		Entry("fractional input with 5 digit after dot", testCase{
			input:    util_proto.Double(50.12345),
			expected: &envoy_type.FractionalPercent{Numerator: 50123450, Denominator: envoy_type.FractionalPercent_MILLION},
		}),
		Entry("fractional input with 7 digit after dot, last digit less than 5", testCase{
			input:    util_proto.Double(50.1234561),
			expected: &envoy_type.FractionalPercent{Numerator: 50123456, Denominator: envoy_type.FractionalPercent_MILLION},
		}),
		Entry("fractional input with 7 digit after dot, last digit more than 5", testCase{
			input:    util_proto.Double(50.1234567),
			expected: &envoy_type.FractionalPercent{Numerator: 50123457, Denominator: envoy_type.FractionalPercent_MILLION},
		}),
	)
})

var _ = Describe("ConvertBandwidth", func() {
	type testCase struct {
		input    string
		expected uint64
	}
	DescribeTable("should properly converts to kbps from gbps, mbps, kbps",
		func(given testCase) {
			// when
			limitKbps, err := ConvertBandwidthToKbps(given.input)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(limitKbps).To(Equal(given.expected))
		},
		Entry("kbps input", testCase{
			input:    "120 kbps",
			expected: 120,
		}),
		Entry("mbps input", testCase{
			input:    "120 mbps",
			expected: 120000,
		}),
		Entry("gbps input", testCase{
			input:    "120 gbps",
			expected: 120000000,
		}),
	)
})
