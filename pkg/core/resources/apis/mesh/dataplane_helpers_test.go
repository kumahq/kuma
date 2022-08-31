package mesh_test

import (
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Dataplane", func() {

	Describe("UsesInterface()", func() {

		type testCase struct {
			dataplane string
			address   string
			port      uint32
			expected  bool
		}

		DescribeTable("should correctly determine whether a given (ip, port) endpoint would overshadow one of Dataplane interfaces",
			func(given testCase) {
				// given
				dataplane := NewDataplaneResource()

				// when
				Expect(util_proto.FromYAML([]byte(given.dataplane), dataplane.Spec)).To(Succeed())
				// then
				Expect(dataplane.UsesInterface(net.ParseIP(given.address), given.port)).To(Equal(given.expected))
			},
			Entry("port of an inbound interface is overshadowed (wilcard ip match)", testCase{
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
				address:  "0.0.0.0",
				port:     80,
				expected: true,
			}),
			Entry("port of the application is overshadowed (wilcard ip match)", testCase{
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
				address:  "0.0.0.0",
				port:     8080,
				expected: true,
			}),
			Entry("port of an outbound interface is overshadowed (wilcard ip match)", testCase{
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
				address:  "0.0.0.0",
				port:     54321,
				expected: true,
			}),
			Entry("port of an inbound interface is overshadowed (exact ip match)", testCase{
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
				address:  "192.168.0.1",
				port:     80,
				expected: true,
			}),
			Entry("port of the application is overshadowed (exact ip match)", testCase{
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
				address:  "192.168.0.1",
				port:     8080,
				expected: true,
			}),
			Entry("port of an outbound interface is overshadowed (exact ip match)", testCase{
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
				address:  "127.0.0.1",
				port:     54321,
				expected: true,
			}),
			Entry("port of invalid inbound interface is not overshadowed", testCase{
				dataplane: `
                networking:
                  inbound:
                   - interface: ?:80:8080
                  outbound:
                  - port: 54321
                    service: db
                  - port: 59200
                    service: elastic
`,
				address:  "0.0.0.0",
				port:     80,
				expected: false,
			}),
			Entry("port of invalid outbound interface is not overshadowed", testCase{
				dataplane: `
                networking:
                  address: 192.168.0.1
                  inbound:
                  - port: 80
                    servicePort: 8080
                  outbound:
                  - interface: ?:54321
                    service: db
                  - port: 59200
                    service: elastic
`,
				address:  "0.0.0.0",
				port:     54321,
				expected: false,
			}),
			Entry("non-overlapping ports are not overshadowed", testCase{
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
				address:  "0.0.0.0",
				port:     5670,
				expected: false,
			}),
			Entry("non-overlapping ip addresses are not overshadowed (inbound listener port)", testCase{
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
				address:  "192.168.0.2",
				port:     80,
				expected: false,
			}),
			Entry("non-overlapping ip addresses are not overshadowed (application port)", testCase{
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
				address:  "192.168.0.2",
				port:     8080,
				expected: false,
			}),
			Entry("non-overlapping ip addresses are not overshadowed (outbound listener port)", testCase{
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
				address:  "192.168.0.2",
				port:     54321,
				expected: false,
			}),
		)

		It("should not crash if Dataplane is nil", func() {
			// given
			var dataplane *DataplaneResource
			address := net.ParseIP("0.0.0.0")
			port := uint32(5670)
			expected := false
			// expect
			Expect(dataplane.UsesInterface(address, port)).To(Equal(expected))
		})
	})

	Describe("GetPrometheusConfig()", func() {

		type testCase struct {
			dataplaneName string
			dataplaneMesh string
			dataplaneSpec string
			meshName      string
			meshSpec      string
			expected      *mesh_proto.PrometheusMetricsBackendConfig
		}

		DescribeTable("should correctly determine effective Prometheus config for given Dataplane and Mesh",
			func(given testCase) {
				// given
				dataplane := NewDataplaneResource()
				if given.dataplaneName != "" {
					dataplane = &DataplaneResource{
						Meta: &test_model.ResourceMeta{
							Name: given.dataplaneName,
							Mesh: given.dataplaneMesh,
						},
						Spec: &mesh_proto.Dataplane{},
					}
					Expect(util_proto.FromYAML([]byte(given.dataplaneSpec), dataplane.Spec)).To(Succeed())
				}

				// given
				var mesh *MeshResource
				if given.meshName != "" {
					mesh = &MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: given.meshName,
						},
						Spec: &mesh_proto.Mesh{},
					}
					Expect(util_proto.FromYAML([]byte(given.meshSpec), mesh.Spec)).To(Succeed())
				}

				// then
				endpoint, err := dataplane.GetPrometheusConfig(mesh)
				Expect(err).ToNot(HaveOccurred())
				Expect(endpoint).To(MatchProto(given.expected))
			},
			Entry("dataplane == `nil` && mesh == `nil`", testCase{
				expected: nil,
			}),
			Entry("dataplane != `nil` && mesh == `nil`", testCase{
				dataplaneName: "backend-01",
				dataplaneSpec: `
                metrics:
                  type: prometheus
                  conf:
                    port: 8765
                    path: /even-more-non-standard-path
`,
				expected: nil,
			}),
			Entry("dataplane.mesh != mesh", testCase{
				dataplaneName: "backend-01",
				dataplaneMesh: "default",
				meshName:      "demo",
				meshSpec: `
                metrics:
                  enabledBackend: prometheus-1
                  backends:
                  - name: prometheus-1
                    type: prometheus
                    conf:
                      port: 1234
                      path: /non-standard-path
`,
				expected: nil,
			}),
			Entry("dataplane.mesh == mesh && mesh.metrics.prometheus == nil", testCase{
				dataplaneName: "backend-01",
				dataplaneMesh: "demo",
				dataplaneSpec: `
                metrics:
                  type: prometheus
                  conf:
                    port: 8765
                    path: /even-more-non-standard-path
`,
				meshName: "demo",
				expected: nil,
			}),
			Entry("dataplane.mesh == mesh && dataplane.metrics.prometheus == nil && mesh.metrics.prometheus != nil", testCase{
				dataplaneName: "backend-01",
				dataplaneMesh: "demo",
				meshName:      "demo",
				meshSpec: `
                metrics:
                  enabledBackend: prometheus-1
                  backends:
                  - name: prometheus-1
                    type: prometheus
                    conf:
                      port: 1234
                      path: /non-standard-path
`,
				expected: &mesh_proto.PrometheusMetricsBackendConfig{
					Port: 1234,
					Path: "/non-standard-path",
				},
			}),
			Entry("dataplane.mesh == mesh && dataplane.metrics.prometheus != nil && mesh.metrics.prometheus != nil", testCase{
				dataplaneName: "backend-01",
				dataplaneMesh: "demo",
				dataplaneSpec: `
                metrics:
                  type: prometheus
                  conf:
                    port: 8765
                    path: /even-more-non-standard-path
`,
				meshName: "demo",
				meshSpec: `
                metrics:
                  enabledBackend: prometheus-1
                  backends:
                  - name: prometheus-1
                    type: prometheus
                    conf:
                      port: 1234
                      path: /non-standard-path
`,
				expected: &mesh_proto.PrometheusMetricsBackendConfig{
					Port: 8765,
					Path: "/even-more-non-standard-path",
				},
			}),
			Entry("dataplane.mesh == mesh && dataplane.metrics.prometheus.conf.aggregate != nil && mesh.metrics.prometheus.conf.aggregate != nil", testCase{
				dataplaneName: "backend-01",
				dataplaneMesh: "demo",
				dataplaneSpec: `
                metrics:
                  type: prometheus
                  conf:
                    port: 8765
                    path: /even-more-non-standard-path
                    aggregate:
                    - name: app1
                      port: 123
                      path: "/stats"
                      enabled: false
`,
				meshName: "demo",
				meshSpec: `
                metrics:
                  enabledBackend: prometheus-1
                  backends:
                  - name: prometheus-1
                    type: prometheus
                    conf:
                      port: 1234
                      path: /non-standard-path
                      aggregate:
                      - name: app1
                        port: 12345
                        path: "/disabled"
`,
				expected: &mesh_proto.PrometheusMetricsBackendConfig{
					Port: 8765,
					Path: "/even-more-non-standard-path",
					Aggregate: []*mesh_proto.PrometheusAggregateMetricsConfig{
						{
							Name:    "app1",
							Port:    123,
							Path:    "/stats",
							Enabled: util_proto.Bool(false),
						},
					},
				},
			}),
		)
	})

	Describe("GetIP()", func() {

		type testCase struct {
			dataplane string
			expected  string
		}

		DescribeTable("should correctly determine IP for a given Dataplane",
			func(given testCase) {
				// given
				var dataplane *DataplaneResource
				if given.dataplane != "" {
					dataplane = NewDataplaneResource()
					Expect(util_proto.FromYAML([]byte(given.dataplane), dataplane.Spec)).To(Succeed())
				}

				// expect
				Expect(dataplane.GetIP()).To(Equal(given.expected))
			},
			Entry("`nil` dataplane", testCase{
				dataplane: ``,
				expected:  "",
			}),
			Entry("dataplane without inbound interfaces", testCase{
				dataplane: `
                networking: {}
`,
				expected: "",
			}),
			Entry("dataplane with address in networking", testCase{
				dataplane: `
                networking:
                  address: 192.168.0.1
                  inbound:
                  - port: 8080
                    address: 192.168.0.2
                    tags:
                      kuma.io/service: backend
`,
				expected: "192.168.0.1",
			}),
			Entry("dataplane with invalid inbound interface", testCase{
				dataplane: `
                networking:
                  inbound:
                  - interface: x.y.z.0
                    tags:
                      kuma.io/service: backend-https
                  - interface: 192.168.0.1:80:8080
                    tags:
                      kuma.io/service: backend
`,
				expected: "",
			}),
		)
	})

	Describe("IsIPv6()", func() {

		type testCase struct {
			dataplane string
			expected  bool
		}

		DescribeTable("should correctly determine IP for a given Dataplane",
			func(given testCase) {
				// given
				var dataplane *DataplaneResource
				if given.dataplane != "" {
					dataplane = NewDataplaneResource()
					Expect(util_proto.FromYAML([]byte(given.dataplane), dataplane.Spec)).To(Succeed())
				}

				// expect
				Expect(dataplane.IsIPv6()).To(Equal(given.expected))
			},
			Entry("`nil` dataplane", testCase{
				dataplane: ``,
				expected:  false,
			}),
			Entry("dataplane without inbound interfaces", testCase{
				dataplane: `
                networking: {}
`,
				expected: false,
			}),
			Entry("dataplane with IPv4 address in networking", testCase{
				dataplane: `
                networking:
                  address: 192.168.0.1
                  inbound:
                  - port: 8080
                    address: 192.168.0.2
                    tags:
                      kuma.io/service: backend
`,
				expected: false,
			}),
			Entry("dataplane with IPv6 address in networking", testCase{
				dataplane: `
                networking:
                  address: fd00::123
                  inbound:
                  - port: 8080
                    address: 192.168.0.2
                    tags:
                      kuma.io/service: backend
`,
				expected: true,
			}),
			Entry("dataplane with invalid inbound interface", testCase{
				dataplane: `
                networking:
                  inbound:
                  - interface: x.y.z.0
                    tags:
                      kuma.io/service: backend-https
                  - interface: 192.168.0.1:80:8080
                    tags:
                      kuma.io/service: backend
`,
				expected: false,
			}),
		)
	})

	Describe("IsUsingTransparentProxy()", func() {
		type testCase struct {
			dataplane string
			expected  bool
		}

		DescribeTable("should correctly determine if dataplane is using transparent proxy",
			func(given testCase) {
				// given
				var dataplane *DataplaneResource
				if given.dataplane != "" {
					dataplane = NewDataplaneResource()
					Expect(util_proto.FromYAML([]byte(given.dataplane), dataplane.Spec)).To(Succeed())
				}

				// expect
				Expect(dataplane.IsUsingTransparentProxy()).To(Equal(given.expected))
			},
			Entry("`nil` dataplane", testCase{
				dataplane: ``,
				expected:  false,
			}),
			Entry("dataplane without transparent proxy", testCase{
				dataplane: `
                networking: {}
`,
				expected: false,
			}),
			Entry("dataplane with empty transparent proxy", testCase{
				dataplane: `
                networking:
                  transparent_proxying: {}
`,
				expected: false,
			}),
			Entry("dataplane with transparent proxy configured", testCase{
				dataplane: `
                networking:
                  transparent_proxying:
                    redirect_port_inbound: 123
                    redirect_port_outbound: 1234
`,
				expected: true,
			}),
			Entry("dataplane with transparent proxy configured and ipv6", testCase{
				dataplane: `
                networking:
                  address: fd00::123
                  transparent_proxying:
                    redirect_port_inbound: 123
                    redirect_port_outbound: 1234
                    redirect_port_inbound_v6: 12345
`,
				expected: true,
			}),
			Entry("dataplane with transparent proxy configured and ipv6 but no port", testCase{
				dataplane: `
                networking:
                  address: fd00::123
                  transparent_proxying:
                    redirect_port_inbound: 123
                    redirect_port_outbound: 1234
`,
				expected: false,
			}),
		)
	})

	var _ = Describe("ParseProtocol()", func() {

		type testCase struct {
			tag      string
			expected Protocol
		}

		DescribeTable("should parse protocol from a tag",
			func(given testCase) {
				Expect(ParseProtocol(given.tag)).To(Equal(given.expected))
			},
			Entry("http", testCase{
				tag:      "http",
				expected: ProtocolHTTP,
			}),
			Entry("tcp", testCase{
				tag:      "tcp",
				expected: ProtocolTCP,
			}),
			Entry("http2", testCase{
				tag:      "http2",
				expected: ProtocolHTTP2,
			}),
			Entry("grpc", testCase{
				tag:      "grpc",
				expected: ProtocolGRPC,
			}),
			Entry("kafka", testCase{
				tag:      "kafka",
				expected: ProtocolKafka,
			}),
			Entry("mongo", testCase{
				tag:      "mongo",
				expected: ProtocolUnknown,
			}),
			Entry("mysql", testCase{
				tag:      "mysql",
				expected: ProtocolUnknown,
			}),
			Entry("unknown", testCase{
				tag:      "unknown",
				expected: ProtocolUnknown,
			}),
			Entry("empty", testCase{
				tag:      "",
				expected: ProtocolUnknown,
			}),
		)
	})
})
