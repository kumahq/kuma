package generator_test

import (
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/golang/protobuf/ptypes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	model "github.com/Kong/kuma/pkg/core/xds"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/generator"
)

var _ = Describe("ProxyTemplateProfileSource", func() {

	type testCase struct {
		mesh            string
		dataplane       string
		profile         string
		envoyConfigFile string
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// setup
			gen := &generator.ProxyTemplateProfileSource{
				ProfileName: given.profile,
			}

			// given
			ctx := xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					SdsLocation: "kuma-system:5677",
					SdsTlsCert:  []byte("12345"),
				},
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
					},
				},
			}

			Expect(util_proto.FromYAML([]byte(given.mesh), &ctx.Mesh.Resource.Spec)).To(Succeed())

			dataplane := mesh_proto.Dataplane{}
			Expect(util_proto.FromYAML([]byte(given.dataplane), &dataplane)).To(Succeed())
			proxy := &model.Proxy{
				Id: model.ProxyId{Name: "demo.backend-01"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name:    "backend-01",
						Mesh:    "demo",
						Version: "1",
					},
					Spec: dataplane,
				},
				TrafficRoutes: model.RouteMap{
					"db": &mesh_core.TrafficRouteResource{
						Spec: mesh_proto.TrafficRoute{
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
								Weight:      100,
								Destination: mesh_proto.MatchService("db"),
							}},
						},
					},
					"elastic": &mesh_core.TrafficRouteResource{
						Spec: mesh_proto.TrafficRoute{
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
								Weight:      100,
								Destination: mesh_proto.MatchService("elastic"),
							}},
						},
					},
				},
				OutboundTargets: model.EndpointMap{
					"db": []model.Endpoint{
						{Target: "192.168.0.3", Port: 5432, Tags: map[string]string{"service": "db", "role": "master"}},
					},
					"elastic": []model.Endpoint{
						{Target: "192.168.0.4", Port: 9200, Tags: map[string]string{"service": "elastic"}},
					},
				},
				HealthChecks: model.HealthCheckMap{
					"elastic": &mesh_core.HealthCheckResource{
						Spec: mesh_proto.HealthCheck{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "elastic"}},
							},
							Conf: &mesh_proto.HealthCheck_Conf{
								ActiveChecks: &mesh_proto.HealthCheck_Conf_Active{
									Interval:           ptypes.DurationProto(5 * time.Second),
									Timeout:            ptypes.DurationProto(4 * time.Second),
									UnhealthyThreshold: 3,
									HealthyThreshold:   2,
								},
							},
						},
					},
				},
				Metadata: &model.DataplaneMetadata{
					AdminPort: 9902,
				},
			}

			// when
			rs, err := gen.Generate(ctx, proxy)

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

			expected, err := ioutil.ReadFile(filepath.Join("testdata", "profile-source", given.envoyConfigFile))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(expected))
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
                    service: backend
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
`,
			profile:         mesh_core.ProfileDefaultProxy,
			envoyConfigFile: "1-envoy-config.golden.yaml",
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
                    service: backend
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
              transparentProxying:
                redirectPort: 15001
`,
			profile:         mesh_core.ProfileDefaultProxy,
			envoyConfigFile: "2-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=false; prometheus_metrics=true", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
            metrics:
              prometheus:
                port: 1234
                path: /non-standard-path
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    service: backend
                    protocol: http
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
`,
			profile:         mesh_core.ProfileDefaultProxy,
			envoyConfigFile: "3-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=true; prometheus_metrics=true", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
            metrics:
              prometheus:
                port: 1234
                path: /non-standard-path
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    service: backend
                    protocol: http
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
              transparentProxying:
                redirectPort: 15001
`,
			profile:         mesh_core.ProfileDefaultProxy,
			envoyConfigFile: "4-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=false; prometheus_metrics=true; prometheus_port=inbound_listener_port", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
            metrics:
              prometheus:
                port: 80
                path: /non-standard-path
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
`,
			profile: mesh_core.ProfileDefaultProxy,
			// we do want to reuse golden file from use case #1 to ensure that
			// Prometheus endpoint does not overshadow port of inbound listener
			envoyConfigFile: "1-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=true; prometheus_metrics=true; prometheus_port=inbound_listener_port", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
            metrics:
              prometheus:
                port: 80
                path: /non-standard-path
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
              transparentProxying:
                redirectPort: 15001
`,
			profile: mesh_core.ProfileDefaultProxy,
			// we do want to reuse golden file from use case #2 to ensure that
			// Prometheus endpoint does not overshadow port of inbound listener
			envoyConfigFile: "2-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=false; prometheus_metrics=true; prometheus_port=application_port", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
            metrics:
              prometheus:
                port: 8080
                path: /non-standard-path
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
`,
			profile: mesh_core.ProfileDefaultProxy,
			// we do want to reuse golden file from use case #1 to ensure that
			// Prometheus endpoint does not overshadow application port
			envoyConfigFile: "1-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=true; prometheus_metrics=true; prometheus_port=application_port", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
            metrics:
              prometheus:
                port: 8080
                path: /non-standard-path
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
              transparentProxying:
                redirectPort: 15001
`,
			profile: mesh_core.ProfileDefaultProxy,
			// we do want to reuse golden file from use case #1 to ensure that
			// Prometheus endpoint does not overshadow application port
			envoyConfigFile: "2-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=false; prometheus_metrics=true; prometheus_port=outbound_listener_port", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
            metrics:
              prometheus:
                port: 54321
                path: /non-standard-path
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
`,
			profile: mesh_core.ProfileDefaultProxy,
			// we do want to reuse golden file from use case #1 to ensure that
			// Prometheus endpoint does not overshadow port of outbound listener
			envoyConfigFile: "1-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=true; prometheus_metrics=true; prometheus_port=outbound_listener_port", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
            metrics:
              prometheus:
                port: 59200
                path: /non-standard-path
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
              transparentProxying:
                redirectPort: 15001
`,
			profile: mesh_core.ProfileDefaultProxy,
			// we do want to reuse golden file from use case #2 to ensure that
			// Prometheus endpoint does not overshadow port of outbound listener
			envoyConfigFile: "2-envoy-config.golden.yaml",
		}),
	)
})
