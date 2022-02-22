package v3_test

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/generator"
	modifications "github.com/kumahq/kuma/pkg/xds/generator/modifications/v3"
)

var _ = Describe("Network Filter modifications", func() {

	type testCase struct {
		listeners     []string
		modifications []string
		expected      string
	}

	DescribeTable("should apply modifications",
		func(given testCase) {
			// given
			set := core_xds.NewResourceSet()
			for _, listenerYAML := range given.listeners {
				listener := &envoy_listener.Listener{}
				err := util_proto.FromYAML([]byte(listenerYAML), listener)
				Expect(err).ToNot(HaveOccurred())
				set.Add(&core_xds.Resource{
					Name:     listener.Name,
					Origin:   generator.OriginInbound,
					Resource: listener,
				})
			}

			var mods []*mesh_proto.ProxyTemplate_Modifications
			for _, modificationYAML := range given.modifications {
				modification := &mesh_proto.ProxyTemplate_Modifications{}
				err := util_proto.FromYAML([]byte(modificationYAML), modification)
				Expect(err).ToNot(HaveOccurred())
				mods = append(mods, modification)
			}

			// when
			err := modifications.Apply(set, mods)

			// then
			Expect(err).ToNot(HaveOccurred())
			resp, err := set.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("should not add filter when there is no filter match", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080`,
			},
			modifications: []string{`
                networkFilter:
                   operation: addFirst
                   value: |
                     name: envoy.filters.network.tcp_proxy
                     typedConfig:
                       '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                       cluster: backend
`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8080
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080
                name: inbound:192.168.0.1:8080`,
		}),
		Entry("should add filter as a first", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: "xyz"`,
			},
			modifications: []string{`
                networkFilter:
                   operation: addFirst
                   value: |
                     name: envoy.filters.network.tcp_proxy
                     typedConfig:
                       '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                       cluster: backend
`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8080
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - filters:
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: backend
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: xyz
                name: inbound:192.168.0.1:8080`,
		}),
		Entry("should remove all filters from all listeners when there is no match section", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: "xyz"
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: backend`,
				`
                name: inbound:192.168.0.1:8081
                filterChains:
                - filters:
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: backend`,
			},
			modifications: []string{`
                networkFilter:
                   operation: remove
`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8080
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - {}
                name: inbound:192.168.0.1:8080
            - name: inbound:192.168.0.1:8081
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - {}
                name: inbound:192.168.0.1:8081`,
		}),
		Entry("should remove all filters from all listeners when there is inbound match section", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: "xyz"
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: backend`,
				`
                name: inbound:192.168.0.1:8081
                filterChains:
                - filters:
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: backend`,
			},
			modifications: []string{`
                networkFilter:
                   operation: remove
                   match:
                     origin: inbound
`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8080
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - {}
                name: inbound:192.168.0.1:8080
            - name: inbound:192.168.0.1:8081
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - {}
                name: inbound:192.168.0.1:8081`,
		}),
		Entry("should remove all filters from picked listener", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: "xyz"
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: backend`,
				`
                name: inbound:192.168.0.1:8081
                filterChains:
                - filters:
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: backend`,
			},
			modifications: []string{`
                networkFilter:
                   operation: remove
                   match:
                     listenerName: inbound:192.168.0.1:8080
`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8080
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - {}
                name: inbound:192.168.0.1:8080
            - name: inbound:192.168.0.1:8081
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - filters:
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: backend
                name: inbound:192.168.0.1:8081`,
		}),
		Entry("should remove all filters of given name from all listeners", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: "xyz"
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: backend`,
				`
                name: inbound:192.168.0.1:8081
                filterChains:
                - filters:
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: backend`,
			},
			modifications: []string{`
                networkFilter:
                   operation: remove
                   match:
                     name: envoy.filters.network.tcp_proxy
`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8080
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - filters:
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: xyz
                name: inbound:192.168.0.1:8080
            - name: inbound:192.168.0.1:8081
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - {}
                name: inbound:192.168.0.1:8081`,
		}),
		Entry("should add filter after already defined (last)", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: "xyz"`,
			},
			modifications: []string{`
                networkFilter:
                   operation: addAfter
                   match:
                     name: envoy.filters.network.direct_response
                   value: |
                     name: envoy.filters.network.tcp_proxy
                     typedConfig:
                       '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                       cluster: backend
`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8080
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - filters:
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: xyz
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: backend
                name: inbound:192.168.0.1:8080`,
		}),
		Entry("should add filter after already defined", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: "xyz"
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                       '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                       cluster: backend
`,
			},
			modifications: []string{`
                networkFilter:
                   operation: addAfter
                   match:
                     name: envoy.filters.network.direct_response
                   value: |
                     name: envoy.filters.network.echo
                     typedConfig:
                       '@type': type.googleapis.com/envoy.extensions.filters.network.echo.v3.Echo
`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8080
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - filters:
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: xyz
                  - name: envoy.filters.network.echo
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.echo.v3.Echo
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: backend
                name: inbound:192.168.0.1:8080`,
		}),
		Entry("should not add filter when name is not matched", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                filterChains:
                - {}`,
			},
			modifications: []string{`
                networkFilter:
                   operation: addAfter
                   match:
                     name: envoy.filters.network.direct_response
                   value: |
                     name: envoy.filters.network.tcp_proxy
                     typedConfig:
                       '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                       cluster: backend
`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8080
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - {}
                name: inbound:192.168.0.1:8080`,
		}),
		Entry("should add filter before already defined (first)", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: "xyz"`,
			},
			modifications: []string{`
                networkFilter:
                   operation: addBefore
                   match:
                     name: envoy.filters.network.direct_response
                   value: |
                     name: envoy.filters.network.tcp_proxy
                     typedConfig:
                       '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                       cluster: backend
`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8080
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - filters:
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: backend
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: xyz
                name: inbound:192.168.0.1:8080`,
		}),
		Entry("should add filter before already defined", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: "xyz"
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                       '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                       cluster: backend
`,
			},
			modifications: []string{`
                networkFilter:
                   operation: addBefore
                   match:
                     name: envoy.filters.network.tcp_proxy
                   value: |
                     name: envoy.filters.network.echo
                     typedConfig:
                       '@type': type.googleapis.com/envoy.extensions.filters.network.echo.v3.Echo
`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8080
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - filters:
                  - name: envoy.filters.network.direct_response
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
                      response:
                        inlineString: xyz
                  - name: envoy.filters.network.echo
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.echo.v3.Echo
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: backend
                name: inbound:192.168.0.1:8080`,
		}),
		Entry("should patch resource matching filter name", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.http_connection_manager
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                      stat_prefix: backend
                      rds:
                        configSource:
                          ads: {}
                        routeConfigName: outbound:backend
                      httpFilters:
                      - name: router
`,
			},
			modifications: []string{`
               networkFilter:
                 operation: patch
                 match:
                   name: envoy.filters.network.http_connection_manager
                 value: |
                   name: envoy.filters.network.http_connection_manager
                   typedConfig:
                     '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                     streamIdleTimeout: 5s
                     requestTimeout: 2s
                     drainTimeout: 10s`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8080
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - filters:
                  - name: envoy.filters.network.http_connection_manager
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                      drainTimeout: 10s
                      httpFilters:
                      - name: router
                      rds:
                        configSource:
                          ads: {}
                        routeConfigName: outbound:backend
                      requestTimeout: 2s
                      statPrefix: backend
                      streamIdleTimeout: 5s
                name: inbound:192.168.0.1:8080`,
		}),
		Entry("should patch resource providing config", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.http_connection_manager
`,
			},
			modifications: []string{`
               networkFilter:
                 operation: patch
                 match:
                   name: envoy.filters.network.http_connection_manager
                 value: |
                   name: envoy.filters.network.http_connection_manager
                   typedConfig:
                     '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                     statPrefix: backend`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8080
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                filterChains:
                - filters:
                  - name: envoy.filters.network.http_connection_manager
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                      statPrefix: backend
                name: inbound:192.168.0.1:8080`,
		}),
	)
})
