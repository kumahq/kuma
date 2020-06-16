package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	model "github.com/Kong/kuma/pkg/core/xds"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/generator"

	test_model "github.com/Kong/kuma/pkg/test/resources/model"
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
					},
				},
			}

			// when
			rs, err := gen.Generate(ctx, given.proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := model.ResourceList(rs).ToDeltaDiscoveryResponse()
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
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
								RedirectPort: 15001,
							},
						},
					},
				},
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
			expected: `
        resources:
        - name: catch_all
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            trafficDirection: OUTBOUND
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
            lbPolicy: CLUSTER_PROVIDED
            name: pass_through
            type: ORIGINAL_DST
          version: v1
`,
		}),
		Entry("transparent_proxying=true with logs", testCase{
			proxy: &model.Proxy{
				Id: model.ProxyId{Name: "side-car"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "v1",
					},
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
								RedirectPort: 15001,
							},
						},
					},
				},
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
			expected: `
        resources:
        - name: catch_all
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Listener
            trafficDirection: OUTBOUND
            address:
              socketAddress:
                address: 0.0.0.0
                portValue: 15001
            filterChains:
            - filters:
              - name: envoy.tcp_proxy
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
                  accessLog:
                  - name: envoy.file_access_log
                    typedConfig:
                      '@type': type.googleapis.com/envoy.config.accesslog.v2.FileAccessLog
                      format: |
                        [%START_TIME%] %RESPONSE_FLAGS% default (unknown)->%UPSTREAM_HOST%(external) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes
                      path: /var/log
                  cluster: pass_through
                  statPrefix: pass_through
            name: catch_all
            useOriginalDst: true
          version: v1
        - name: pass_through
          resource:
            '@type': type.googleapis.com/envoy.api.v2.Cluster
            connectTimeout: 5s
            lbPolicy: CLUSTER_PROVIDED
            name: pass_through
            type: ORIGINAL_DST
          version: v1
`,
		}),
	)
})
