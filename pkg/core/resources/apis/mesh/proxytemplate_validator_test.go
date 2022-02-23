package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	_ "github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ = Describe("ProxyTemplate", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(spec string) {
				proxyTemplate := mesh.NewProxyTemplateResource()

				// when
				err := util_proto.FromYAML([]byte(spec), proxyTemplate.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				err = proxyTemplate.Validate()
				// then
				Expect(err).ToNot(HaveOccurred())
			},
			Entry("full example", `
                selectors:
                - match:
                    kuma.io/service: backend
                conf:
                  imports:
                  - default-proxy
                  resources:
                  - name: additional
                    version: v1
                    resource: | 
                      '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
                      connection_pool_per_downstream_connection: true # V3 only setting
                      connectTimeout: 5s
                      loadAssignment:
                        clusterName: localhost:8443
                        endpoints:
                          - lbEndpoints:
                              - endpoint:
                                  address:
                                    socketAddress:
                                      address: 127.0.0.1
                                      portValue: 8443
                      name: localhost:8443
                      type: STATIC`,
			),
			Entry("empty conf", `
                selectors:
                - match:
                    kuma.io/service: backend`,
			),
			Entry("cluster modifications", `
                selectors:
                - match:
                    kuma.io/service: backend
                conf:
                  modifications:
                  - cluster:
                      operation: add
                      value: |
                        name: xyz
                        connectTimeout: 5s
                        type: STATIC
                  - cluster:
                      operation: patch
                      value: |
                        connectTimeout: 5s
                  - cluster:
                      operation: patch
                      match:
                        name: inbound:127.0.0.1:8080
                      value: |
                        connectTimeout: 5s
                  - cluster:
                      operation: patch
                      match:
                        origin: inbound
                      value: |
                        connectTimeout: 5s
                  - cluster:
                      operation: remove
                  - cluster:
                      operation: remove
                      match:
                        name: inbound:127.0.0.1:8080
                  - cluster:
                      operation: remove
                      match:
                        origin: inbound
                  `,
			),
			Entry("listener modifications", `
                selectors:
                - match:
                    kuma.io/service: backend
                conf:
                  modifications:
                  - listener:
                      operation: add
                      value: |
                        name: xyz
                        address:
                          socketAddress:
                            address: 192.168.0.1
                            portValue: 8080
                  - listener:
                      operation: patch
                      value: |
                        address:
                          socketAddress:
                            portValue: 8080
                  - listener:
                      operation: patch
                      match:
                        name: inbound:127.0.0.1:8080
                      value: |
                        address:
                          socketAddress:
                            portValue: 8080
                  - listener:
                      operation: patch
                      match:
                        origin: inbound
                      value: |
                        address:
                          socketAddress:
                            portValue: 8080
                  - listener:
                      operation: remove
                  - listener:
                      operation: remove
                      match:
                        name: inbound:127.0.0.1:8080
                  - listener:
                      operation: remove
                      match:
                        origin: inbound
                  `,
			),
			Entry("network filter modifications", `
                selectors:
                - match:
                    kuma.io/service: backend
                conf:
                  modifications:
                  - networkFilter:
                      operation: addFirst
                      value: |
                        name: envoy.filters.network.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                          cluster: backend
                  - networkFilter:
                      operation: addLast
                      value: |
                        name: envoy.filters.network.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                          cluster: backend
                  - networkFilter:
                      operation: addBefore
                      match:
                        name: envoy.filters.network.direct_response
                      value: |
                        name: envoy.filters.network.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                          cluster: backend
                  - networkFilter:
                      operation: addAfter
                      match:
                        name: envoy.filters.network.direct_response
                        listenerName: inbound:127.0.0.0:8080
                      value: |
                        name: envoy.filters.network.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                          cluster: backend
                  - networkFilter:
                      operation: patch
                      match:
                        name: envoy.filters.network.tcp_proxy
                        listenerName: inbound:127.0.0.0:8080
                      value: |
                        name: envoy.filters.network.tcp_proxy
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                          cluster: backend
                  - networkFilter:
                      operation: remove
                  `,
			),
			Entry("http filter modifications", `
                selectors:
                - match:
                    kuma.io/service: backend
                conf:
                  modifications:
                  - httpFilter:
                      operation: addFirst
                      value: |
                        name: envoy.filters.http.router
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                          dynamicStats: false
                  - httpFilter:
                      operation: addLast
                      value: |
                        name: envoy.filters.http.router
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                          dynamicStats: false
                  - httpFilter:
                      operation: addAfter
                      match:
                        name: envoy.filters.http.router
                      value: |
                        name: envoy.filters.http.gzip
                  - httpFilter:
                      operation: addAfter
                      match:
                        name: envoy.filters.http.router
                        listenerName: inbound:127.0.0.0:8080
                      value: |
                        name: envoy.filters.http.gzip
                  - httpFilter:
                      operation: patch
                      match:
                        name: envoy.filters.network.tcp_proxy
                        listenerName: inbound:127.0.0.0:8080
                      value: |
                        name: envoy.filters.http.router
                        typedConfig:
                          '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                          dynamicStats: false
                  - httpFilter:
                      operation: remove
                  `,
			),
			Entry("virtual host modifications", `
                selectors:
                - match:
                    kuma.io/service: backend
                conf:
                  modifications:
                  - virtualHost:
                      operation: add
                      match:
                        origin: outbound
                        routeConfigurationName: outbound:backend
                      value: |
                        name: backend
                        domains:
                        - backend.com
                        routes:
                        - match:
                            prefix: /
                          route:
                            cluster: backend
                  - virtualHost:
                      operation: patch
                      value: |
                        retryPolicy:
                          retryOn: 5xx
                          numRetries: 3
                  - virtualHost:
                      operation: patch
                      match:
                        origin: outbound
                      value: |
                        retryPolicy:
                          retryOn: 5xx
                          numRetries: 3
                  - virtualHost:
                      operation: patch
                      match:
                        routeConfigurationName: outbound:backend
                        name: backend
                      value: |
                        retryPolicy:
                          retryOn: 5xx
                          numRetries: 3
                  - cluster:
                      operation: remove
                  - cluster:
                      operation: remove
                      match:
                        routeConfigurationName: outbound:backend
                        name: backend
                  - cluster:
                      operation: remove
                      match:
                        origin: inbound
                  `,
			),
		)

		type testCase struct {
			proxyTemplate string
			expected      string
		}
		DescribeTable("should validate fields",
			func(given testCase) {
				// given
				proxyTemplate := mesh.NewProxyTemplateResource()

				// when
				err := util_proto.FromYAML([]byte(given.proxyTemplate), proxyTemplate.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := proxyTemplate.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty import", testCase{
				proxyTemplate: `
                conf:
                  imports:
                  - ""
                selectors:
                - match:
                    kuma.io/service: backend`,
				expected: `
                violations:
                - field: conf.imports[0]
                  message: cannot be empty`,
			}),
			Entry("unknown profile", testCase{
				proxyTemplate: `
                conf:
                  imports:
                  - unknown-profile
                selectors:
                - match:
                    kuma.io/service: backend`,
				expected: `
                violations:
                - field: conf.imports[0]
                  message: 'profile not found. Available profiles: default-proxy'`,
			}),
			Entry("resources empty fields", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    kuma.io/service: backend
                conf:
                  resources:
                  - name:
                    version:
                    resource:`,
				expected: `
                violations:
                - field: conf.resources[0].name
                  message: cannot be empty
                - field: conf.resources[0].version
                  message: cannot be empty
                - field: conf.resources[0].resource
                  message: cannot be empty`,
			}),
			Entry("selector without tags", testCase{
				proxyTemplate: `
                selectors:
                - match:`,
				expected: `
                violations:
                - field: selectors[0].match
                  message: must have at least one tag
                - field: selectors[0].match
                  message: mandatory tag "kuma.io/service" is missing`,
			}),
			Entry("empty tag", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    "": asdf`,
				expected: `
                violations:
                - field: selectors[0].match
                  message: tag name must be non-empty
                - field: selectors[0].match
                  message: mandatory tag "kuma.io/service" is missing`,
			}),
			Entry("empty tag value", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    kuma.io/service:`,
				expected: `
                violations:
                - field: 'selectors[0].match["kuma.io/service"]'
                  message: tag value must be non-empty`,
			}),
			Entry("validation error from envoy protobuf resource", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    kuma.io/service: backend
                conf:
                  resources:
                  - name: additional
                    version: v1
                    resource: | 
                      '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
                      loadAssignment:
                        clusterName: localhost:8443
                        endpoints:
                          - lbEndpoints:
                              - endpoint:
                                  address:
                                    socketAddress:
                                      address: 127.0.0.1
                                      portValue: 8443`,
				expected: `
                violations:
                - field: conf.resources[0].resource
                  message: 'native Envoy resource is not valid: invalid Cluster.Name: value length must be at least 1 runes'`,
			}),
			Entry("invalid envoy resource", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    kuma.io/service: backend
                conf:
                  resources:
                  - name: additional
                    version: v1
                    resource: not-envoy-resource`,
				expected: `
                violations:
                - field: conf.resources[0].resource
                  message: 'native Envoy resource is not valid: json: cannot unmarshal string into Go value of type map[string]json.RawMessage'`,
			}),
			Entry("invalid cluster modifications", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    kuma.io/service: backend
                conf:
                  modifications:
                  - cluster:
                      operation: addFirst
                  - cluster:
                      operation: add
                      value: '{'
                  - cluster:
                      operation: patch
                      value: '{'
                  - cluster:
                      operation: add
                      match:
                        name: inbound:127.0.0.1:80
                      value: |
                        name: xyz`,
				expected: `
                violations:
                - field: conf.modifications[0].cluster.operation
                  message: 'invalid operation. Available operations: "add", "patch", "remove"'
                - field: conf.modifications[1].cluster.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: conf.modifications[2].cluster.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: conf.modifications[3].cluster.match
                  message: cannot be defined`,
			}),
			Entry("invalid listener modifications", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    kuma.io/service: backend
                conf:
                  modifications:
                  - listener:
                      operation: addFirst
                  - listener:
                      operation: add
                      value: '{'
                  - listener:
                      operation: patch
                      value: '{'
                  - listener:
                      operation: add
                      match:
                        name: inbound:127.0.0.1:80
                      value: |
                        name: xyz
                        address:
                          socketAddress:
                            address: 192.168.0.1
                            portValue: 8080`,
				expected: `
                violations:
                - field: conf.modifications[0].listener.operation
                  message: 'invalid operation. Available operations: "add", "patch", "remove"'
                - field: conf.modifications[1].listener.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: conf.modifications[2].listener.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: conf.modifications[3].listener.match
                  message: cannot be defined`,
			}),
			Entry("invalid network filter operation", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    kuma.io/service: backend
                conf:
                  modifications:
                  - networkFilter:
                      operation: addFirst
                      value: '{'
                  - networkFilter:
                      operation: addBefore
                      value: '{'
                  - networkFilter:
                      operation: addAfter
                      value: '{'
                  - networkFilter:
                      operation: patch
                      value: '{'
`,
				expected: `
                violations:
                - field: conf.modifications[0].networkFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: conf.modifications[1].networkFilter.match.name
                  message: cannot be empty. You need to pick a filter before which this one will be added
                - field: conf.modifications[1].networkFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: conf.modifications[2].networkFilter.match.name
                  message: cannot be empty. You need to pick a filter after which this one will be added
                - field: conf.modifications[2].networkFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: conf.modifications[3].networkFilter.match.name
                  message: cannot be empty
                - field: conf.modifications[3].networkFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'`,
			}),
			Entry("invalid http filter operation", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    kuma.io/service: backend
                conf:
                  modifications:
                  - httpFilter:
                      operation: addFirst
                      value: '{'
                  - httpFilter:
                      operation: addBefore
                      value: '{'
                  - httpFilter:
                      operation: addAfter
                      value: '{'
                  - httpFilter:
                      operation: patch
                      value: '{'
`,
				expected: `
                violations:
                - field: conf.modifications[0].httpFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: conf.modifications[1].httpFilter.match.name
                  message: cannot be empty. You need to pick a filter before which this one will be added
                - field: conf.modifications[1].httpFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: conf.modifications[2].httpFilter.match.name
                  message: cannot be empty. You need to pick a filter after which this one will be added
                - field: conf.modifications[2].httpFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: conf.modifications[3].httpFilter.match.name
                  message: cannot be empty
                - field: conf.modifications[3].httpFilter.value
                  message: 'native Envoy resource is not valid: unexpected EOF'`,
			}),
			Entry("invalid virtual host operation", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    kuma.io/service: backend
                conf:
                  modifications:
                  - virtualHost:
                      operation: add
                      match:
                        name: xyz
                      value: '{'
                  - virtualHost:
                      operation: addFirst
                  - virtualHost:
                      operation: patch
                      value: '{'
`,
				expected: `
                violations:
                - field: conf.modifications[0].virtualHost.match.name
                  message: cannot be defined
                - field: conf.modifications[0].virtualHost.value
                  message: 'native Envoy resource is not valid: unexpected EOF'
                - field: conf.modifications[1].virtualHost.operation
                  message: 'invalid operation. Available operations: "add", "patch", "remove"'
                - field: conf.modifications[2].virtualHost.value
                  message: 'native Envoy resource is not valid: unexpected EOF'`,
			}),
		)
	})
})
