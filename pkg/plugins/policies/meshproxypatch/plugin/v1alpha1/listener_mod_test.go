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

var _ = Describe("Listener modifications", func() {
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
					Origin:   metadata.OriginInbound,
					Resource: listener,
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
		Entry("should add listener", testCase{
			modifications: []string{
				`
                listener:
                   operation: Add
                   value: |
                     name: inbound:192.168.0.1:8080
                     trafficDirection: INBOUND
                     address:
                       socketAddress:
                         address: 192.168.0.1
                         portValue: 8080
                     filterChains:
                     - filters:
                       - name: envoy.filters.network.tcp_proxy
                         typedConfig:
                           '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                           cluster: localhost:8080
                           statPrefix: localhost_8080`,
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
                filterChains:
                - filters:
                  - name: envoy.filters.network.tcp_proxy
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
                      cluster: localhost:8080
                      statPrefix: localhost_8080
                name: inbound:192.168.0.1:8080
                trafficDirection: INBOUND`,
		}),
		Entry("should replace listener", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                trafficDirection: INBOUND
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080`,
			},
			modifications: []string{
				`
                listener:
                   operation: Add
                   value: |
                     name: inbound:192.168.0.1:8080
                     address:
                       socketAddress:
                         address: 192.168.0.2
                         portValue: 8090`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8080
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                name: inbound:192.168.0.1:8080
                address:
                  socketAddress:
                    address: 192.168.0.2
                    portValue: 8090`,
		}),
		Entry("should remove listener matching all", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                trafficDirection: INBOUND
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080`,
			},
			modifications: []string{
				`
                listener:
                   operation: Remove`,
			},
			expected: `{}`,
		}),
		Entry("should remove listener matching name", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                trafficDirection: INBOUND
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080`,
				`
                name: inbound:192.168.0.1:8081
                trafficDirection: INBOUND
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8081`,
			},
			modifications: []string{
				`
                listener:
                   operation: Remove
                   match:
                     name: inbound:192.168.0.1:8080`,
			},
			expected: `
            resources:
            - name: inbound:192.168.0.1:8081
              resource:
                '@type': type.googleapis.com/envoy.config.listener.v3.Listener
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8081
                name: inbound:192.168.0.1:8081
                trafficDirection: INBOUND`,
		}),
		Entry("should remove all inbound listeners", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                trafficDirection: INBOUND
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080`,
			},
			modifications: []string{
				`
                listener:
                   operation: Remove
                   match:
                     origin: inbound`,
			},
			expected: `{}`,
		}),
		Entry("should patch listener matching name", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                trafficDirection: INBOUND
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080`,
			},
			modifications: []string{
				`
                listener:
                   operation: Patch
                   match:
                     name: inbound:192.168.0.1:8080
                   value: |
                     tcpFastOpenQueueLength: 32`,
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
                name: inbound:192.168.0.1:8080
                tcpFastOpenQueueLength: 32
                trafficDirection: INBOUND`,
		}),
		Entry("should patch listener matching name with JsonPatch", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                trafficDirection: INBOUND
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080
                enableReusePort: true
                tcpBacklogSize: 256
                `,
			},
			modifications: []string{
				`
                listener:
                   operation: Patch
                   match:
                     name: inbound:192.168.0.1:8080
                   jsonPatches:
                   - op: add
                     path: /tcpFastOpenQueueLength
                     value: 88
                   - op: replace
                     path: /enableReusePort
                     value: false
                   - op: remove
                     path: /tcpBacklogSize
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
                enableReusePort: false
                name: inbound:192.168.0.1:8080
                tcpFastOpenQueueLength: 88
                trafficDirection: INBOUND
            `,
		}),
		Entry("should patch listener matching metadata", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                trafficDirection: INBOUND
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080
                metadata:
                  filterMetadata:
                    io.kuma.tags:
                      kuma.io/service: backend`,
			},
			modifications: []string{
				`
                listener:
                   operation: Patch
                   match:
                     name: inbound:192.168.0.1:8080
                     tags:
                       kuma.io/service: backend
                   value: |
                     tcpFastOpenQueueLength: 32`,
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
                metadata:
                  filterMetadata:
                    io.kuma.tags:
                      kuma.io/service: backend
                name: inbound:192.168.0.1:8080
                tcpFastOpenQueueLength: 32
                trafficDirection: INBOUND`,
		}),
		Entry("should not patch listener with non-matching metadata", testCase{
			listeners: []string{
				`
                name: inbound:192.168.0.1:8080
                trafficDirection: INBOUND
                address:
                  socketAddress:
                    address: 192.168.0.1
                    portValue: 8080
                metadata:
                  filterMetadata:
                    io.kuma.tags:
                      kuma.io/service: backend`,
			},
			modifications: []string{
				`
                listener:
                   operation: Patch
                   match:
                     name: inbound:192.168.0.1:8080
                     tags:
                       kuma.io/service: web
                   value: |
                     tcpFastOpenQueueLength: 32`,
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
                metadata:
                  filterMetadata:
                    io.kuma.tags:
                      kuma.io/service: backend
                name: inbound:192.168.0.1:8080
                trafficDirection: INBOUND`,
		}),
	)
})
