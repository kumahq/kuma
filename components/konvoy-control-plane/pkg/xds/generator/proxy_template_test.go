package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/model/api/v1alpha1"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/generator"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/model"
)

var _ = Describe("Generator", func() {
	Describe("TransparentInboundProxyProfile", func() {

		type testCase struct {
			proxy    *model.Proxy
			expected string
		}

		DescribeTable("Generate Snapshot per Envoy Node",
			func(given testCase) {
				// setup
				gen := &generator.TransparentInboundProxyProfile{}

				// when
				rs, err := gen.Generate(given.proxy)

				// then
				Expect(err).ToNot(HaveOccurred())

				// then
				resp := generator.ResourceList(rs).ToDeltaDiscoveryResponse()
				actual, err := util_proto.ToYAML(resp)

				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("should support Nodes without IP addresses and Ports", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version: "v1",
					},
				},
				expected: `{}`,
			}),
			Entry("should support Nodes with IP addresses but without Ports", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1"},
					},
				},
				expected: `{}`,
			}),
			Entry("should support Nodes with 1 IP address and 1 Port", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1"},
						Ports:     []uint32{8080},
					},
				},
				expected: `
        resources:
        - name: localhost:8080
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            connectTimeout: 5s
            loadAssignment:
              clusterName: localhost:8080
              endpoints:
              - lbEndpoints:
                - endpoint:
                    address:
                      socketAddress:
                        address: 127.0.0.1
                        portValue: 8080
            name: localhost:8080
            type: STATIC
          version: v1
        - name: inbound:192.168.0.1:8080
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            deprecatedV1:
              bindToPort: false
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: localhost:8080
                  statPrefix: localhost:8080
            name: inbound:192.168.0.1:8080
          version: v1
`,
			}),
			Entry("should support Nodes with 1 IP address and 2 Ports", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1"},
						Ports:     []uint32{8080, 8443},
					},
				},
				expected: `
        resources:
          - name: localhost:8080
            resource:
              '@type': type.googleapis.com/envoy.api.v2.Cluster
              connectTimeout: 5s
              loadAssignment:
                clusterName: localhost:8080
                endpoints:
                - lbEndpoints:
                  - endpoint:
                      address:
                        socketAddress:
                          address: 127.0.0.1
                          portValue: 8080
              name: localhost:8080
              type: STATIC
            version: v1
          - name: inbound:192.168.0.1:8080
            resource:
              '@type': type.googleapis.com/envoy.api.v2.Listener
              address:
                socketAddress:
                  address: 192.168.0.1
                  portValue: 8080
              deprecatedV1:
                bindToPort: false
              filterChains:
              - filters:
                - name: envoy.tcp_proxy
                  typedConfig:
                    '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                    cluster: localhost:8080
                    statPrefix: localhost:8080
              name: inbound:192.168.0.1:8080
            version: v1
          - name: localhost:8443
            resource:
              '@type': type.googleapis.com/envoy.api.v2.Cluster
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
              type: STATIC
            version: v1
          - name: inbound:192.168.0.1:8443
            resource:
              '@type': type.googleapis.com/envoy.api.v2.Listener
              address:
                socketAddress:
                  address: 192.168.0.1
                  portValue: 8443
              deprecatedV1:
                bindToPort: false
              filterChains:
              - filters:
                - name: envoy.tcp_proxy
                  typedConfig:
                    '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                    cluster: localhost:8443
                    statPrefix: localhost:8443
              name: inbound:192.168.0.1:8443
            version: v1
`,
			}),
			Entry("should support Nodes with 2 IP addresses and 2 Ports", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1", "192.168.0.2"},
						Ports:     []uint32{8080, 8443},
					},
				},
				expected: `
        resources:
        - name: localhost:8080
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            connectTimeout: 5s
            loadAssignment:
              clusterName: localhost:8080
              endpoints:
              - lbEndpoints:
                - endpoint:
                    address:
                      socketAddress:
                        address: 127.0.0.1
                        portValue: 8080
            name: localhost:8080
            type: STATIC
          version: v1
        - name: inbound:192.168.0.1:8080
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            deprecatedV1:
              bindToPort: false
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: localhost:8080
                  statPrefix: localhost:8080
            name: inbound:192.168.0.1:8080
          version: v1
        - name: inbound:192.168.0.2:8080
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            address:
              socketAddress:
                address: 192.168.0.2
                portValue: 8080
            deprecatedV1:
              bindToPort: false
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: localhost:8080
                  statPrefix: localhost:8080
            name: inbound:192.168.0.2:8080
          version: v1
        - name: localhost:8443
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
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
            type: STATIC
          version: v1
        - name: inbound:192.168.0.1:8443
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8443
            deprecatedV1:
              bindToPort: false
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: localhost:8443
                  statPrefix: localhost:8443
            name: inbound:192.168.0.1:8443
          version: v1
        - name: inbound:192.168.0.2:8443
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            address:
              socketAddress:
                address: 192.168.0.2
                portValue: 8443
            deprecatedV1:
              bindToPort: false
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: localhost:8443
                  statPrefix: localhost:8443
            name: inbound:192.168.0.2:8443
          version: v1
`,
			}),
		)
	})

	Describe("TransparentOutboundProxyProfile", func() {

		type testCase struct {
			proxy    *model.Proxy
			expected string
		}

		DescribeTable("Generate Snapshot per Envoy Node",
			func(given testCase) {
				// setup
				gen := &generator.TransparentOutboundProxyProfile{}

				// when
				rs, err := gen.Generate(given.proxy)

				// then
				Expect(err).ToNot(HaveOccurred())

				// then
				resp := generator.ResourceList(rs).ToDeltaDiscoveryResponse()
				actual, err := util_proto.ToYAML(resp)

				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("should support Nodes without IP addresses and ports", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version: "v1",
					},
				},
				expected: `
        resources:
        - name: catch_all
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            address:
              socketAddress:
                address: 0.0.0.0
                portValue: 15001
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: pass_through
                  statPrefix: pass_through
            name: catch_all
            useOriginalDst: true
          version: v1
        - name: pass_through
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            connectTimeout: 5s
            lbPolicy: ORIGINAL_DST_LB
            name: pass_through
            type: ORIGINAL_DST
          version: v1
`,
			}),
			Entry("should support Nodes with 1 IP address and 1 Port", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1"},
						Ports:     []uint32{8080},
					},
				},
				expected: `
        resources:
        - name: catch_all
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            address:
              socketAddress:
                address: 0.0.0.0
                portValue: 15001
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: pass_through
                  statPrefix: pass_through
            name: catch_all
            useOriginalDst: true
          version: v1
        - name: pass_through
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            connectTimeout: 5s
            lbPolicy: ORIGINAL_DST_LB
            name: pass_through
            type: ORIGINAL_DST
          version: v1
`,
			}),
			Entry("should support Nodes with 2 IP addresses and 2 Ports", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1", "192.168.0.2"},
						Ports:     []uint32{8080, 8443},
					},
				},
				expected: `
        resources:
        - name: catch_all
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            address:
              socketAddress:
                address: 0.0.0.0
                portValue: 15001
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: pass_through
                  statPrefix: pass_through
            name: catch_all
            useOriginalDst: true
          version: v1
        - name: pass_through
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            connectTimeout: 5s
            lbPolicy: ORIGINAL_DST_LB
            name: pass_through
            type: ORIGINAL_DST
          version: v1
`,
			}),
		)
	})

	Describe("ProxyTemplateProfileSource", func() {

		type testCase struct {
			proxy    *model.Proxy
			profile  *konvoy_mesh.ProxyTemplateProfileSource
			expected string
		}

		DescribeTable("Generate Snapshot per Envoy Node",
			func(given testCase) {
				// setup
				gen := &generator.ProxyTemplateProfileSource{
					Profile: given.profile,
				}

				// when
				rs, err := gen.Generate(given.proxy)

				// then
				Expect(err).ToNot(HaveOccurred())

				// then
				resp := generator.ResourceList(rs).ToDeltaDiscoveryResponse()
				actual, err := util_proto.ToYAML(resp)

				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("should support pre-defined `transparent-inbound-proxy` profile", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1"},
						Ports:     []uint32{8080},
					},
				},
				profile: &konvoy_mesh.ProxyTemplateProfileSource{
					Name: generator.ProfileTransparentInboundProxy,
				},
				expected: `
        resources:
          - name: localhost:8080
            resource:
              '@type': type.googleapis.com/envoy.api.v2.Cluster
              connectTimeout: 5s
              loadAssignment:
                clusterName: localhost:8080
                endpoints:
                - lbEndpoints:
                  - endpoint:
                      address:
                        socketAddress:
                          address: 127.0.0.1
                          portValue: 8080
              name: localhost:8080
              type: STATIC
            version: v1
          - name: inbound:192.168.0.1:8080
            resource:
              '@type': type.googleapis.com/envoy.api.v2.Listener
              address:
                socketAddress:
                  address: 192.168.0.1
                  portValue: 8080
              deprecatedV1:
                bindToPort: false
              filterChains:
              - filters:
                - name: envoy.tcp_proxy
                  typedConfig:
                    '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                    cluster: localhost:8080
                    statPrefix: localhost:8080
              name: inbound:192.168.0.1:8080
            version: v1
`,
			}),
			Entry("should support pre-defined `transparent-outbound-proxy` profile", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1"},
						Ports:     []uint32{8080},
					},
				},
				profile: &konvoy_mesh.ProxyTemplateProfileSource{
					Name: "transparent-outbound-proxy",
				},
				expected: `
        resources:
          - name: catch_all
            resource:
              '@type': type.googleapis.com/envoy.api.v2.Listener
              address:
                socketAddress:
                  address: 0.0.0.0
                  portValue: 15001
              filterChains:
              - filters:
                - name: envoy.tcp_proxy
                  typedConfig:
                    '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                    cluster: pass_through
                    statPrefix: pass_through
              name: catch_all
              useOriginalDst: true
            version: v1
          - name: pass_through
            resource:
              '@type': type.googleapis.com/envoy.api.v2.Cluster
              connectTimeout: 5s
              lbPolicy: ORIGINAL_DST_LB
              name: pass_through
              type: ORIGINAL_DST
            version: v1
`,
			}),
		)
	})

	Describe("ProxyTemplateRawSource", func() {

		type testCase struct {
			proxy    *model.Proxy
			raw      *konvoy_mesh.ProxyTemplateRawSource
			expected string
		}

		DescribeTable("Generate Snapshot per Envoy Node",
			func(given testCase) {
				// setup
				gen := &generator.ProxyTemplateRawSource{
					Raw: given.raw,
				}

				// when
				rs, err := gen.Generate(given.proxy)

				// then
				Expect(err).ToNot(HaveOccurred())

				// then
				resp := generator.ResourceList(rs).ToDeltaDiscoveryResponse()
				actual, err := util_proto.ToYAML(resp)

				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("should support empty resource list", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1"},
						Ports:     []uint32{8080},
					},
				},
				raw: &konvoy_mesh.ProxyTemplateRawSource{
					Resources: nil,
				},
				expected: `{}`,
			}),
			Entry("should support Listener resource as YAML", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1"},
						Ports:     []uint32{8080},
					},
				},
				raw: &konvoy_mesh.ProxyTemplateRawSource{
					Resources: []konvoy_mesh.ProxyTemplateRawResource{{
						Name:    "raw-name",
						Version: "raw-version",
						Resource: `
            '@type': type.googleapis.com/envoy.api.v2.Listener
            address:
              socketAddress:
                address: 0.0.0.0
                portValue: 15001
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: pass_through
                  statPrefix: pass_through
            name: catch_all
            useOriginalDst: true
`,
					}},
				},
				expected: `
        resources:
          - name: raw-name
            resource:
              '@type': type.googleapis.com/envoy.api.v2.Listener
              address:
                socketAddress:
                  address: 0.0.0.0
                  portValue: 15001
              filterChains:
              - filters:
                - name: envoy.tcp_proxy
                  typedConfig:
                    '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                    cluster: pass_through
                    statPrefix: pass_through
              name: catch_all
              useOriginalDst: true
            version: raw-version
`,
			}),
			Entry("should support Cluster resource as YAML", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1"},
						Ports:     []uint32{8080},
					},
				},
				raw: &konvoy_mesh.ProxyTemplateRawSource{
					Resources: []konvoy_mesh.ProxyTemplateRawResource{{
						Name:    "raw-name",
						Version: "raw-version",
						Resource: `
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            connectTimeout: 5s
            loadAssignment:
              clusterName: localhost:8080
              endpoints:
              - lbEndpoints:
                - endpoint:
                    address:
                      socketAddress:
                        address: 127.0.0.1
                        portValue: 8080
            name: localhost:8080
            type: STATIC
`,
					}},
				},
				expected: `
        resources:
          - name: raw-name
            resource:
              '@type': type.googleapis.com/envoy.api.v2.Cluster
              connectTimeout: 5s
              loadAssignment:
                clusterName: localhost:8080
                endpoints:
                - lbEndpoints:
                  - endpoint:
                      address:
                        socketAddress:
                          address: 127.0.0.1
                          portValue: 8080
              name: localhost:8080
              type: STATIC
            version: raw-version
`,
			}),
			Entry("should support Cluster resource as JSON", testCase{
				proxy: &model.Proxy{
					Id: model.ProxyId{Name: "side-car", Namespace: "default"},
					Workload: model.Workload{
						Version:   "v1",
						Addresses: []string{"192.168.0.1"},
						Ports:     []uint32{8080},
					},
				},
				raw: &konvoy_mesh.ProxyTemplateRawSource{
					Resources: []konvoy_mesh.ProxyTemplateRawResource{{
						Name:    "raw-name",
						Version: "raw-version",
						Resource: `
            {
              "@type": "type.googleapis.com/envoy.api.v2.Cluster",
              "connectTimeout": "5s",
              "loadAssignment": {
                "clusterName": "localhost:8080",
                "endpoints": [
                  {
                    "lbEndpoints": [
                      {
                        "endpoint": {
                          "address": {
                            "socketAddress": {
                              "address": "127.0.0.1",
                              "portValue": 8080
                            }
                          }
                        }
                      }
                    ]
                  }
                ]
              },
              "name": "localhost:8080",
              "type": "STATIC"
            }
`,
					}},
				},
				expected: `
        resources:
          - name: raw-name
            resource:
              '@type': type.googleapis.com/envoy.api.v2.Cluster
              connectTimeout: 5s
              loadAssignment:
                clusterName: localhost:8080
                endpoints:
                - lbEndpoints:
                  - endpoint:
                      address:
                        socketAddress:
                          address: 127.0.0.1
                          portValue: 8080
              name: localhost:8080
              type: STATIC
            version: raw-version
`,
			}),
		)
	})
})
