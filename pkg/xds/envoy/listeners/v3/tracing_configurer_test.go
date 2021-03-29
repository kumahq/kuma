package v3_test

import (
	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/kumahq/kuma/pkg/core/xds"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("TracingConfigurer", func() {

	type testCase struct {
		backend  *mesh_proto.TracingBackend
		expected string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder(envoy.APIV3).
				Configure(InboundListener("inbound:192.168.0.1:8080", "192.168.0.1", 8080, xds.SocketAddressProtocolTCP)).
				Configure(FilterChain(NewFilterChainBuilder(envoy.APIV3).
					Configure(HttpConnectionManager("localhost:8080")).
					Configure(Tracing(given.backend)))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("backend specified with sampling", testCase{
			backend: &mesh_proto.TracingBackend{
				Name:     "zipkin",
				Sampling: &wrappers.DoubleValue{Value: 30.5},
				Type:     mesh_proto.TracingZipkinType,
				Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
					Url: "http://zipkin.us:9090/v2/spans",
				}),
			},
			expected: `
            address:
              socketAddress:
                address: 192.168.0.1
                ipv4Compat: true
                portValue: 8080
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  httpFilters:
                  - name: envoy.filters.http.router
                  statPrefix: localhost_8080
                  tracing:
                    overallSampling:
                      value: 30.5
                    provider:
                      name: envoy.zipkin
                      typedConfig:
                        '@type': type.googleapis.com/envoy.config.trace.v3.ZipkinConfig
                        collectorCluster: tracing:zipkin
                        collectorEndpoint: /v2/spans
                        collectorEndpointVersion: HTTP_JSON
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND`,
		}),
		Entry("backend specified without sampling", testCase{
			backend: &mesh_proto.TracingBackend{
				Name: "zipkin",
				Type: mesh_proto.TracingZipkinType,
				Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
					Url: "http://zipkin.us:9090/v2/spans",
				}),
			},
			expected: `
            address:
              socketAddress:
                address: 192.168.0.1
                ipv4Compat: true
                portValue: 8080
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  httpFilters:
                  - name: envoy.filters.http.router
                  statPrefix: localhost_8080
                  tracing:
                    provider:
                      name: envoy.zipkin
                      typedConfig:
                        '@type': type.googleapis.com/envoy.config.trace.v3.ZipkinConfig
                        collectorCluster: tracing:zipkin
                        collectorEndpoint: /v2/spans
                        collectorEndpointVersion: HTTP_JSON
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND`,
		}),
		Entry("no backend specified", testCase{
			backend: nil,
			expected: `
            name: inbound:192.168.0.1:8080
            trafficDirection: INBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                ipv4Compat: true
                portValue: 8080
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  statPrefix: localhost_8080
                  httpFilters:
                  - name: envoy.filters.http.router
`,
		}),
	)
})
