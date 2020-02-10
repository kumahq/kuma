package listeners_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/envoy/listeners"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"

	util_error "github.com/Kong/kuma/pkg/util/error"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("UpdateFilterConfig()", func() {

	Context("happy path", func() {
		type testCase struct {
			listener   *v2.Listener
			filterName string
			updateFunc func(proto.Message) error
			expected   string
		}

		DescribeTable("should update filter config",
			func(given testCase) {
				// when
				err := UpdateFilterConfig(given.listener, given.filterName, given.updateFunc)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				actual, err := util_proto.ToYAML(given.listener)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("0 chains", testCase{
				listener:   &v2.Listener{},
				filterName: envoy_wellknown.TCPProxy,
				updateFunc: func(proto.Message) error { return errors.New("should never happen") },
				expected:   `{}`,
			}),
			Entry("1 chain, 0 filters", testCase{
				listener: &v2.Listener{
					FilterChains: []*envoy_listener.FilterChain{{}},
				},
				filterName: envoy_wellknown.TCPProxy,
				updateFunc: func(proto.Message) error { return errors.New("should never happen") },
				expected: `
                filterChains:
                - {}
`,
			}),
			Entry("1 chain, 1 filter", func() testCase {
				pbst, err := ptypes.MarshalAny(&envoy_tcp.TcpProxy{})
				util_error.MustNot(err)
				return testCase{
					listener: &v2.Listener{
						FilterChains: []*envoy_listener.FilterChain{{
							Filters: []*envoy_listener.Filter{{
								Name: envoy_wellknown.TCPProxy,
								ConfigType: &envoy_listener.Filter_TypedConfig{
									TypedConfig: pbst,
								},
							}},
						}},
					},
					filterName: envoy_wellknown.TCPProxy,
					updateFunc: func(filterConfig proto.Message) error {
						proxy := filterConfig.(*envoy_tcp.TcpProxy)
						proxy.ClusterSpecifier = &envoy_tcp.TcpProxy_Cluster{
							Cluster: "backend",
						}
						return nil
					},
					expected: `
                    filterChains:
                    - filters:
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: backend
`,
				}
			}()),
			Entry("1 chain, 2 filters", func() testCase {
				pbst, err := ptypes.MarshalAny(&envoy_tcp.TcpProxy{})
				util_error.MustNot(err)
				return testCase{
					listener: &v2.Listener{
						FilterChains: []*envoy_listener.FilterChain{{
							Filters: []*envoy_listener.Filter{
								{
									Name: envoy_wellknown.RoleBasedAccessControl,
								},
								{
									Name: envoy_wellknown.TCPProxy,
									ConfigType: &envoy_listener.Filter_TypedConfig{
										TypedConfig: pbst,
									},
								},
							},
						}},
					},
					filterName: envoy_wellknown.TCPProxy,
					updateFunc: func(filterConfig proto.Message) error {
						proxy := filterConfig.(*envoy_tcp.TcpProxy)
						proxy.ClusterSpecifier = &envoy_tcp.TcpProxy_Cluster{
							Cluster: "backend",
						}
						return nil
					},
					expected: `
                    filterChains:
                    - filters:
                      - name: envoy.filters.network.rbac
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: backend
`,
				}
			}()),
			Entry("2 chain, 2 filters", func() testCase {
				pbst, err := ptypes.MarshalAny(&envoy_tcp.TcpProxy{})
				util_error.MustNot(err)
				return testCase{
					listener: &v2.Listener{
						FilterChains: []*envoy_listener.FilterChain{
							{
								Filters: []*envoy_listener.Filter{
									{
										Name: envoy_wellknown.RoleBasedAccessControl,
									},
									{
										Name: envoy_wellknown.TCPProxy,
										ConfigType: &envoy_listener.Filter_TypedConfig{
											TypedConfig: pbst,
										},
									},
								},
							},
							{
								Filters: []*envoy_listener.Filter{
									{
										Name: envoy_wellknown.TCPProxy,
										ConfigType: &envoy_listener.Filter_TypedConfig{
											TypedConfig: pbst,
										},
									},
									{
										Name: envoy_wellknown.TCPProxy,
										ConfigType: &envoy_listener.Filter_TypedConfig{
											TypedConfig: pbst,
										},
									},
								},
							},
						},
					},
					filterName: envoy_wellknown.TCPProxy,
					updateFunc: func(filterConfig proto.Message) error {
						proxy := filterConfig.(*envoy_tcp.TcpProxy)
						proxy.ClusterSpecifier = &envoy_tcp.TcpProxy_Cluster{
							Cluster: "backend",
						}
						return nil
					},
					expected: `
                    filterChains:
                    - filters:
                      - name: envoy.filters.network.rbac
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: backend
                    - filters:
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: backend
                      - name: envoy.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                          cluster: backend
`,
				}
			}()),
		)
	})

	Context("error path", func() {

		type testCase struct {
			listener    *v2.Listener
			filterName  string
			updateFunc  func(proto.Message) error
			expectedErr string
		}

		DescribeTable("should return an error",
			func(given testCase) {
				// when
				err := UpdateFilterConfig(given.listener, given.filterName, given.updateFunc)
				// then
				Expect(err).To(HaveOccurred())
				// and
				Expect(err.Error()).To(Equal(given.expectedErr))
			},
			Entry("1 chain, 1 filter without config", testCase{
				listener: &v2.Listener{
					FilterChains: []*envoy_listener.FilterChain{{
						Filters: []*envoy_listener.Filter{{
							Name: envoy_wellknown.TCPProxy,
						}},
					}},
				},
				filterName:  envoy_wellknown.TCPProxy,
				updateFunc:  func(proto.Message) error { return errors.New("should never happen") },
				expectedErr: `filter_chains[0].filters[0]: config cannot be 'nil'`,
			}),
			Entry("1 chain, 1 filter with a wrong config type", func() testCase {
				pbst, err := ptypes.MarshalAny(&envoy_hcm.HttpConnectionManager{})
				util_error.MustNot(err)
				return testCase{
					listener: &v2.Listener{
						FilterChains: []*envoy_listener.FilterChain{{
							Filters: []*envoy_listener.Filter{{
								Name: envoy_wellknown.TCPProxy,
								ConfigType: &envoy_listener.Filter_TypedConfig{
									TypedConfig: pbst,
								},
							}},
						}},
					},
					filterName:  envoy_wellknown.TCPProxy,
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
			Expect(err.Error()).To(Equal(given.expectedErr))
		},
		Entry("TcpProxy instead of HttpConnectionManager", testCase{
			inputActual:   &envoy_tcp.TcpProxy{},
			inputExpected: &envoy_hcm.HttpConnectionManager{},
			expectedErr:   `filter config has unexpected type: expected *envoy_config_filter_network_http_connection_manager_v2.HttpConnectionManager, got *envoy_config_filter_network_tcp_proxy_v2.TcpProxy`,
		}),
	)
})
