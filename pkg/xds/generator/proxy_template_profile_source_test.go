package generator_test

import (
	"context"
	"path/filepath"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/runtime"
	"github.com/kumahq/kuma/pkg/tls"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v3"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

type dummyCLACache struct {
	outboundTargets model.EndpointMap
}

func (d *dummyCLACache) GetCLA(ctx context.Context, meshName, meshHash string, cluster envoy_common.Cluster, version envoy_common.APIVersion) (proto.Message, error) {
	return endpoints.CreateClusterLoadAssignment(cluster.Service(), d.outboundTargets[cluster.Service()]), nil
}

var _ model.CLACache = &dummyCLACache{}

var _ = Describe("ProxyTemplateProfileSource", func() {

	type testCase struct {
		mesh      string
		dataplane string
		profile   string
		expected  string
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// setup
			gen := &generator.ProxyTemplateProfileSource{
				ProfileName: given.profile,
			}

			// given
			outboundTargets := model.EndpointMap{
				"db": []model.Endpoint{
					{
						Target: "192.168.0.3",
						Port:   5432,
						Tags:   map[string]string{"kuma.io/service": "db", "role": "master"},
						Weight: 1,
					},
				},
				"elastic": []model.Endpoint{
					{
						Target: "192.168.0.4",
						Port:   9200,
						Tags:   map[string]string{"kuma.io/service": "elastic"},
						Weight: 1,
					},
				},
			}
			ctx := xds_context.Context{
				ConnectionInfo: xds_context.ConnectionInfo{
					Authority: "kuma-system:5677",
				},
				ControlPlane: &xds_context.ControlPlaneContext{
					SdsTlsCert: []byte("12345"),
					AdminProxyKeyPair: &tls.KeyPair{
						CertPEM: []byte("LS0=="),
						KeyPEM:  []byte("LS0=="),
					},
					CLACache: &dummyCLACache{outboundTargets: outboundTargets},
				},
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
						Spec: &mesh_proto.Mesh{},
					},
				},
				EnvoyAdminClient: &runtime.DummyEnvoyAdminClient{},
			}

			Expect(util_proto.FromYAML([]byte(given.mesh), ctx.Mesh.Resource.Spec)).To(Succeed())

			dataplane := &mesh_proto.Dataplane{}
			Expect(util_proto.FromYAML([]byte(given.dataplane), dataplane)).To(Succeed())

			proxy := &model.Proxy{
				Id: *model.BuildProxyId("", "demo.backend-01"),
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name:    "backend-01",
						Mesh:    "demo",
						Version: "1",
					},
					Spec: dataplane,
				},
				APIVersion: envoy_common.APIV3,
				Routing: model.Routing{
					TrafficRoutes: model.RouteMap{
						mesh_proto.OutboundInterface{
							DataplaneIP:   "127.0.0.1",
							DataplanePort: 54321,
						}: &mesh_core.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Destination: mesh_proto.MatchService("db"),
								},
							},
						},
						mesh_proto.OutboundInterface{
							DataplaneIP:   "127.0.0.1",
							DataplanePort: 59200,
						}: &mesh_core.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Destination: mesh_proto.MatchService("elastic"),
								},
							},
						},
					},
					OutboundTargets: outboundTargets,
				},
				Policies: model.MatchedPolicies{
					HealthChecks: model.HealthCheckMap{
						"elastic": &mesh_core.HealthCheckResource{
							Spec: &mesh_proto.HealthCheck{
								Sources: []*mesh_proto.Selector{
									{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
								},
								Destinations: []*mesh_proto.Selector{
									{Match: mesh_proto.TagSelector{"kuma.io/service": "elastic"}},
								},
								Conf: &mesh_proto.HealthCheck_Conf{
									Interval:           durationpb.New(5 * time.Second),
									Timeout:            durationpb.New(4 * time.Second),
									UnhealthyThreshold: 3,
									HealthyThreshold:   2,
								},
							},
						},
					},
				},
				Metadata: &model.DataplaneMetadata{
					AdminPort: 9902,
					Version: &mesh_proto.Version{
						KumaDp: &mesh_proto.KumaDpVersion{
							Version: "1.2.0",
						},
					},
				},
			}

			// when
			rs, err := gen.Generate(ctx, proxy)

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

			// and output matches golden files
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "profile-source", given.expected)))
		},
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=false", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    kuma.io/service: backend
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
`,
			profile:  mesh_core.ProfileDefaultProxy,
			expected: "1-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=true", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    kuma.io/service: backend
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
              transparentProxying:
                redirectPortOutbound: 15001
                redirectPortInbound: 15006
`,
			profile:  mesh_core.ProfileDefaultProxy,
			expected: "2-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=false; prometheus_metrics=true", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
            metrics:
              enabledBackend: prometheus-1
              backends:
              - name: prometheus-1
                type: prometheus
                conf:
                  port: 1234
                  path: /non-standard-path
                  skipMTLS: false
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    kuma.io/service: backend
                    kuma.io/protocol: http
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
`,
			profile:  mesh_core.ProfileDefaultProxy,
			expected: "3-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=true; prometheus_metrics=true", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
            metrics:
              enabledBackend: prometheus-1
              backends:
              - name: prometheus-1
                type: prometheus
                conf:
                  port: 1234
                  path: /non-standard-path
                  skipMTLS: false
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    kuma.io/service: backend
                    kuma.io/protocol: http
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
              transparentProxying:
                redirectPortOutbound: 15001
                redirectPortInbound: 15006
`,
			profile:  mesh_core.ProfileDefaultProxy,
			expected: "4-envoy-config.golden.yaml",
		}),
	)
})
