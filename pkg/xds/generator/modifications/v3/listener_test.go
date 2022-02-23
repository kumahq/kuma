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
		Entry("should add listener", testCase{
			modifications: []string{`
                listener:
                   operation: add
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
                   operation: add
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
                   operation: remove`,
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
                   operation: remove
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
                   operation: remove
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
                   operation: patch
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
	)
})
