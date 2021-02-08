package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"

	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("TransparentProxyGenerator", func() {

	type testCase struct {
		proxy    *model.Proxy
		expected string
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// setup
			gen := &generator.TransparentProxyGenerator{}
			ctx := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{},
					},
				},
			}

			// when
			rs, err := gen.Generate(ctx, given.proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(resp)
			// then
			Expect(err).ToNot(HaveOccurred())

			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("transparent_proxying=false", testCase{
			proxy: &model.Proxy{
				Id: model.ProxyId{Name: "side-car"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "v1",
					},
				},
				APIVersion: envoy_common.APIV2,
			},
			expected: `
        {}
`,
		}),
		Entry("transparent_proxying=true", testCase{
			proxy: &model.Proxy{
				Id: model.ProxyId{Name: "side-car"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "v1",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
								RedirectPortOutbound: 15001,
								RedirectPortInbound:  15006,
							},
						},
					},
				},
				APIVersion: envoy_common.APIV2,
				Policies: model.MatchedPolicies{
					Logs: map[model.ServiceName]*mesh_proto.LoggingBackend{ // to show that is not picked
						"some-service": {
							Name: "file",
							Type: mesh_proto.LoggingFileType,
							Conf: util_proto.MustToStruct(&mesh_proto.FileLoggingBackendConfig{
								Path: "/var/log",
							}),
						},
					},
				},
			},
			expected: `
        resources:
        - name: inbound:passthrough
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            altStatName: inbound_passthrough
            connectTimeout: 5s
            lbPolicy: CLUSTER_PROVIDED
            name: inbound:passthrough
            type: ORIGINAL_DST
            upstreamBindConfig:
              sourceAddress:
                address: 127.0.0.6
                portValue: 0
        - name: outbound:passthrough
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            altStatName: outbound_passthrough
            connectTimeout: 5s
            lbPolicy: CLUSTER_PROVIDED
            name: outbound:passthrough
            type: ORIGINAL_DST
        - name: inbound:passthrough
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            address:
              socketAddress:
                address: 0.0.0.0
                portValue: 15006
            filterChains:
            - filters:
              - name: envoy.filters.network.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: inbound:passthrough
                  statPrefix: inbound_passthrough
            name: inbound:passthrough
            trafficDirection: INBOUND
            useOriginalDst: true
        - name: outbound:passthrough
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            address:
              socketAddress:
                address: 0.0.0.0
                portValue: 15001
            filterChains:
            - filters:
              - name: envoy.filters.network.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: outbound:passthrough
                  statPrefix: outbound_passthrough
            name: outbound:passthrough
            trafficDirection: OUTBOUND
            useOriginalDst: true
`,
		}),
		Entry("transparent_proxying=true with logs", testCase{
			proxy: &model.Proxy{
				Id: model.ProxyId{Name: "side-car"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "v1",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
								RedirectPortOutbound: 15001,
								RedirectPortInbound:  15006,
							},
						},
					},
				},
				APIVersion: envoy_common.APIV2,
				Policies: model.MatchedPolicies{
					Logs: map[model.ServiceName]*mesh_proto.LoggingBackend{ // to show that is is not picked
						"pass_through": {
							Name: "file",
							Type: mesh_proto.LoggingFileType,
							Conf: util_proto.MustToStruct(&mesh_proto.FileLoggingBackendConfig{
								Path: "/var/log",
							}),
						},
					},
				},
			},
			expected: `
        resources:
        - name: inbound:passthrough
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            altStatName: inbound_passthrough
            connectTimeout: 5s
            lbPolicy: CLUSTER_PROVIDED
            name: inbound:passthrough
            type: ORIGINAL_DST
            upstreamBindConfig:
              sourceAddress:
                address: 127.0.0.6
                portValue: 0
        - name: outbound:passthrough
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            altStatName: outbound_passthrough
            connectTimeout: 5s
            lbPolicy: CLUSTER_PROVIDED
            name: outbound:passthrough
            type: ORIGINAL_DST
        - name: inbound:passthrough
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            address:
              socketAddress:
                address: 0.0.0.0
                portValue: 15006
            filterChains:
            - filters:
              - name: envoy.filters.network.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  cluster: inbound:passthrough
                  statPrefix: inbound_passthrough
            name: inbound:passthrough
            trafficDirection: INBOUND
            useOriginalDst: true
        - name: outbound:passthrough
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            address:
              socketAddress:
                address: 0.0.0.0
                portValue: 15001
            filterChains:
            - filters:
              - name: envoy.filters.network.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  accessLog:
                  - name: envoy.access_loggers.file
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.accesslog.v2.FileAccessLog
                      format: |+
                        [%START_TIME%] %RESPONSE_FLAGS% default (unknown)->%UPSTREAM_HOST%(external) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes
                        
                      path: /var/log
                  cluster: outbound:passthrough
                  statPrefix: outbound_passthrough
            name: outbound:passthrough
            trafficDirection: OUTBOUND
            useOriginalDst: true
`,
		}),
	)
})
