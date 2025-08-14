package v1alpha1_test

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/plugin/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/generator/metadata"
)

var _ = Describe("Virtual Host modifications", func() {
	type testCase struct {
		routeCfgs     []string
		modifications []string
		expected      string
	}

	DescribeTable("should apply modifications",
		func(given testCase) {
			// given
			set := core_xds.NewResourceSet()
			for _, routeCfgYAML := range given.routeCfgs {
				routeCfg := &envoy_listener.Listener{}
				err := util_proto.FromYAML([]byte(routeCfgYAML), routeCfg)
				Expect(err).ToNot(HaveOccurred())
				set.Add(&core_xds.Resource{
					Name:     routeCfg.Name,
					Origin:   metadata.OriginOutbound,
					Resource: routeCfg,
				})
			}

			var mods []api.Modification
			for _, modificationYAML := range given.modifications {
				modification := api.Modification{}
				err := yaml.Unmarshal([]byte(modificationYAML), &modification)
				Expect(err).ToNot(HaveOccurred())
				mods = append(mods, modification)
			}

			// when
			err := plugin.ApplyMods(set, mods)

			// then
			Expect(err).ToNot(HaveOccurred())
			resp, err := set.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("should add virtual host", testCase{
			routeCfgs: []string{
				`
                name: outbound:192.168.0.1:8080
                trafficDirection: INBOUND
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.http_connection_manager
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                      statPrefix: localhost_8080
                      httpFilters:
                      - name: envoy.filters.http.router
                      routeConfig:
                        name: outbound:backend
`,
			},
			modifications: []string{
				`
                virtualHost:
                   operation: Add
                   value: |
                     name: backend
                     domains:
                     - backend.com
                     routes:
                     - match:
                         prefix: /
                       route:
                         cluster: backend`,
			},
			expected: `
                resources:
                - name: outbound:192.168.0.1:8080
                  resource:
                    '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                    address:
                      socketAddress:
                        address: 192.168.0.1
                        portValue: 8080
                    filterChains:
                    - filters:
                      - name: envoy.filters.network.http_connection_manager
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                          httpFilters:
                          - name: envoy.filters.http.router
                          routeConfig:
                            name: outbound:backend
                            virtualHosts:
                            - domains:
                              - backend.com
                              name: backend
                              routes:
                              - match:
                                  prefix: /
                                route:
                                  cluster: backend
                          statPrefix: localhost_8080
                    name: outbound:192.168.0.1:8080
                    trafficDirection: INBOUND
`,
		}),
		Entry("should remove virtual host", testCase{
			routeCfgs: []string{
				`
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.http_connection_manager
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                      httpFilters:
                      - name: envoy.filters.http.router
                      routeConfig:
                        name: outbound:backend
                        virtualHosts:
                        - domains:
                          - backend.com
                          name: backend
                          routes:
                          - match:
                              prefix: /
                            route:
                              cluster: backend
                      statPrefix: localhost_8080
                name: outbound:192.168.0.1:8080
                trafficDirection: INBOUND
`,
			},
			modifications: []string{
				`
                virtualHost:
                   operation: Remove
                   match:
                     name: backend`,
			},
			expected: `
                resources:
                - name: outbound:192.168.0.1:8080
                  resource:
                    '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                    address:
                      socketAddress:
                        address: 192.168.0.1
                        portValue: 8080
                    filterChains:
                    - filters:
                      - name: envoy.filters.network.http_connection_manager
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                          httpFilters:
                          - name: envoy.filters.http.router
                          routeConfig:
                            name: outbound:backend
                          statPrefix: localhost_8080
                    name: outbound:192.168.0.1:8080
                    trafficDirection: INBOUND`,
		}),
		Entry("should remove virtual hosts of given name from all listeners", testCase{
			routeCfgs: []string{
				`
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.http_connection_manager
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                      httpFilters:
                      - name: envoy.filters.http.router
                      routeConfig:
                        name: outbound:backend
                        virtualHosts:
                        - domains:
                          - backend.com
                          name: backend
                          routes:
                          - match:
                              prefix: /
                            route:
                              cluster: backend
                      statPrefix: localhost_8080
                name: outbound:192.168.0.1:8080
                trafficDirection: INBOUND
`,
				`
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 80
                filterChains:
                - filters:
                  - name: envoy.filters.network.http_connection_manager
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                      httpFilters:
                      - name: envoy.filters.http.router
                      routeConfig:
                        name: outbound:backend
                        virtualHosts:
                        - domains:
                          - backend.com
                          name: backend
                          routes:
                          - match:
                              prefix: /
                            route:
                              cluster: backend
                      statPrefix: localhost_80
                name: outbound:192.168.0.1:80
                trafficDirection: INBOUND
`,
			},
			modifications: []string{
				`
                virtualHost:
                   operation: Remove
                   match:
                     name: backend`,
			},
			expected: `
                resources:
                - name: outbound:192.168.0.1:80
                  resource:
                    '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                    address:
                      socketAddress:
                        address: 192.168.0.1
                        portValue: 80
                    filterChains:
                    - filters:
                      - name: envoy.filters.network.http_connection_manager
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                          httpFilters:
                          - name: envoy.filters.http.router
                          routeConfig:
                            name: outbound:backend
                          statPrefix: localhost_80
                    name: outbound:192.168.0.1:80
                    trafficDirection: INBOUND
                - name: outbound:192.168.0.1:8080
                  resource:
                    '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                    address:
                      socketAddress:
                        address: 192.168.0.1
                        portValue: 8080
                    filterChains:
                    - filters:
                      - name: envoy.filters.network.http_connection_manager
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                          httpFilters:
                          - name: envoy.filters.http.router
                          routeConfig:
                            name: outbound:backend
                          statPrefix: localhost_8080
                    name: outbound:192.168.0.1:8080
                    trafficDirection: INBOUND`,
		}),
		Entry("should patch a virtual host", testCase{
			routeCfgs: []string{
				`
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.http_connection_manager
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                      httpFilters:
                      - name: envoy.filters.http.router
                      routeConfig:
                        name: outbound:backend
                        virtualHosts:
                        - domains:
                          - backend.com
                          name: backend
                          routes:
                          - match:
                              prefix: /
                            route:
                              cluster: backend
                      statPrefix: localhost_8080
                name: outbound:192.168.0.1:8080
                trafficDirection: INBOUND
`,
			},
			modifications: []string{
				`
                virtualHost:
                   operation: Patch
                   match:
                     origin: outbound
                   value: |
                     retryPolicy:
                       retryOn: 5xx
                       numRetries: 3`,
			},
			expected: `
                resources:
                - name: outbound:192.168.0.1:8080
                  resource:
                    '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                    address:
                      socketAddress:
                        address: 192.168.0.1
                        portValue: 8080
                    filterChains:
                    - filters:
                      - name: envoy.filters.network.http_connection_manager
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                          httpFilters:
                          - name: envoy.filters.http.router
                          routeConfig:
                            name: outbound:backend
                            virtualHosts:
                            - domains:
                              - backend.com
                              name: backend
                              retryPolicy:
                                numRetries: 3
                                retryOn: 5xx
                              routes:
                              - match:
                                  prefix: /
                                route:
                                  cluster: backend
                          statPrefix: localhost_8080
                    name: outbound:192.168.0.1:8080
                    trafficDirection: INBOUND`,
		}),
		Entry("should patch a virtual host with JsonPatch", testCase{
			routeCfgs: []string{
				`
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.http_connection_manager
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                      httpFilters:
                      - name: envoy.filters.http.router
                      routeConfig:
                        name: outbound:backend
                        virtualHosts:
                        - domains:
                          - backend.com
                          name: backend
                          requestHeadersToRemove:
                          - foo
                          - bar
                          routes:
                          - match:
                              prefix: /
                            route:
                              cluster: backend
                      statPrefix: localhost_8080
                name: outbound:192.168.0.1:8080
                trafficDirection: INBOUND
`,
			},
			modifications: []string{
				`
                virtualHost:
                  operation: Patch
                  match:
                    origin: outbound
                  jsonPatches:
                  - op: remove
                    path: /requestHeadersToRemove/1
                  - op: replace
                    path: /requestHeadersToRemove/0
                    value: baz
                  - op: add
                    path: /retryPolicy
                    value:
                      retryOn: 5xx
                      numRetries: 3
                `,
			},
			expected: `
                resources:
                - name: outbound:192.168.0.1:8080
                  resource:
                    '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                    address:
                      socketAddress:
                        address: 192.168.0.1
                        portValue: 8080
                    filterChains:
                    - filters:
                      - name: envoy.filters.network.http_connection_manager
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                          httpFilters:
                          - name: envoy.filters.http.router
                          routeConfig:
                            name: outbound:backend
                            virtualHosts:
                            - domains:
                              - backend.com
                              name: backend
                              requestHeadersToRemove:
                              - baz
                              retryPolicy:
                                numRetries: 3
                                retryOn: 5xx
                              routes:
                              - match:
                                  prefix: /
                                route:
                                  cluster: backend
                          statPrefix: localhost_8080
                    name: outbound:192.168.0.1:8080
                    trafficDirection: INBOUND`,
		}),
		Entry("should patch a virtual host adding new route", testCase{
			routeCfgs: []string{
				`
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080
                filterChains:
                - filters:
                  - name: envoy.filters.network.http_connection_manager
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                      httpFilters:
                      - name: envoy.filters.http.router
                      routeConfig:
                        name: outbound:backend
                        virtualHosts:
                        - domains:
                          - backend.com
                          name: backend
                          routes:
                          - match:
                              prefix: /
                            route:
                              cluster: backend
                      statPrefix: localhost_8080
                name: outbound:192.168.0.1:8080
                trafficDirection: INBOUND
`,
			},
			modifications: []string{
				`
                virtualHost:
                   operation: Patch
                   match:
                     routeConfigurationName: outbound:backend
                   value: |
                     routes:
                     - match:
                         prefix: /web
                       route:
                         cluster: web`,
			},
			expected: `
                resources:
                - name: outbound:192.168.0.1:8080
                  resource:
                    '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                    address:
                      socketAddress:
                        address: 192.168.0.1
                        portValue: 8080
                    filterChains:
                    - filters:
                      - name: envoy.filters.network.http_connection_manager
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                          httpFilters:
                          - name: envoy.filters.http.router
                          routeConfig:
                            name: outbound:backend
                            virtualHosts:
                            - domains:
                              - backend.com
                              name: backend
                              routes:
                              - match:
                                  prefix: /
                                route:
                                  cluster: backend
                              - match:
                                  prefix: /web
                                route:
                                  cluster: web
                          statPrefix: localhost_8080
                    name: outbound:192.168.0.1:8080
                    trafficDirection: INBOUND`,
		}),
	)
})
