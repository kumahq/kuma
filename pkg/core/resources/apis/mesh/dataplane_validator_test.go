package mesh_test

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/validators"
	"github.com/Kong/kuma/pkg/test/resources/model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dataplane", func() {

	var validDataplane core_mesh.DataplaneResource

	BeforeEach(func() {
		validDataplane = core_mesh.DataplaneResource{
			Meta: &model.ResourceMeta{
				Name: "dp-1",
				Mesh: "mesh-1",
			},
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port: 8080,
							Tags: map[string]string{
								"service": "backend",
								"version": "1",
							},
						},
					},
					Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
						{
							Port:    3333,
							Service: "redis",
						},
					},
				},
			},
		}
	})

	DescribeTable("should pass validation",
		func(dataplaneFn func() core_mesh.DataplaneResource) {
			// given
			dataplane := dataplaneFn()

			// when
			err := dataplane.Validate()

			// then
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("legacy - valid dataplane with inbounds", func() core_mesh.DataplaneResource {
			return core_mesh.DataplaneResource{
				Meta: &model.ResourceMeta{
					Name: "dp-1",
					Mesh: "mesh-1",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Interface: "127.0.0.1:8080:9090",
								Tags: map[string]string{
									"service": "backend",
									"version": "1",
								},
							},
						},
						Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
							{
								Interface: ":3333",
								Service:   "redis",
							},
						},
					},
				},
			}
		}),
		Entry("dataplane with inbounds", func() core_mesh.DataplaneResource {
			return validDataplane
		}),
		Entry("dataplane with inbounds that has service ports", func() core_mesh.DataplaneResource {
			validDataplane.Spec.Networking.Inbound[0].ServicePort = 4567
			return validDataplane
		}),
		Entry("dataplane with inbounds that has an address", func() core_mesh.DataplaneResource {
			validDataplane.Spec.Networking.Inbound[0].Address = "192.168.0.1"
			return validDataplane
		}),
		Entry("dataplane with gateway", func() core_mesh.DataplaneResource {
			return core_mesh.DataplaneResource{
				Meta: &model.ResourceMeta{
					Name: "dp-1",
					Mesh: "mesh-1",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						//Address: "192.168.0.1", to have backwards compatibility we do not validate it yet
						Gateway: &mesh_proto.Dataplane_Networking_Gateway{
							Tags: map[string]string{
								"service": "gateway",
							},
						},
						Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
							{
								Port:    3333,
								Service: "redis",
							},
						},
					},
				},
			}
		}),
	)

	type testCase struct {
		dataplane        func() core_mesh.DataplaneResource
		validationResult *validators.ValidationError
	}
	DescribeTable("should catch validation errors",
		func(given testCase) {
			// when
			dataplane := given.dataplane()
			err := dataplane.Validate()

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(given.validationResult))
		},
		Entry("not enough inbound interfaces and no gateway", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound = []*mesh_proto.Dataplane_Networking_Inbound{}
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   "networking",
						Message: `has to contain at least one inbound interface or gateway`,
					},
				},
			},
		}),
		Entry("both inbounds and gateway are defined", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Gateway = &mesh_proto.Dataplane_Networking_Gateway{
					Tags: map[string]string{
						"service": "gateway",
					},
				}
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   "networking",
						Message: `inbound cannot be defined both with gateway`,
					},
				},
			},
		}),
		Entry("inbound: empty port", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound[0].Port = 0
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.inbound[0].port`,
						Message: `port has to in range of [1, 65535]`,
					},
				},
			},
		}),
		Entry("inbound: port outside of the range", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound[0].Port = 65536
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.inbound[0].port`,
						Message: `port has to in range of [1, 65535]`,
					},
				},
			},
		}),
		Entry("inbound: servicePort outside of the range", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound[0].ServicePort = 65536
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.inbound[0].servicePort`,
						Message: `servicePort has to in range of [1, 65535]`,
					},
				},
			},
		}),
		Entry("inbound: invalid address", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound[0].Address = "invalid"
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.inbound[0].address`,
						Message: `address has to be valid IP address`,
					},
				},
			},
		}),
		Entry("inbound: empty service tag", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound[0].Tags = map[string]string{}
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.inbound[0].tags["service"]`,
						Message: `tag has to exist`,
					},
				},
			},
		}),
		Entry("inbound: empty tag", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound[0].Tags["version"] = ""
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.inbound[0].tags["version"]`,
						Message: `tag value cannot be empty`,
					},
				},
			},
		}),
		Entry("gateway: empty service tag", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound = nil
				validDataplane.Spec.Networking.Gateway = &mesh_proto.Dataplane_Networking_Gateway{
					Tags: map[string]string{},
				}
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.gateway.tags["service"]`,
						Message: `tag has to exist`,
					},
				},
			},
		}),
		Entry("gateway: empty tag", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound = nil
				validDataplane.Spec.Networking.Gateway = &mesh_proto.Dataplane_Networking_Gateway{
					Tags: map[string]string{
						"service": "gateway",
						"version": "",
					},
				}
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.gateway.tags["version"]`,
						Message: `tag value cannot be empty`,
					},
				},
			},
		}),
		Entry("outbound: empty service tag", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Outbound[0].Service = ""
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.outbound[0].service`,
						Message: `cannot be empty`,
					},
				},
			},
		}),
		Entry("outbound: empty port", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Outbound[0].Port = 0
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.outbound[0].port`,
						Message: `port has to in range of [1, 65535]`,
					},
				},
			},
		}),
		Entry("outbound: port outside of the range", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Outbound[0].Port = 65536
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.outbound[0].port`,
						Message: `port has to in range of [1, 65535]`,
					},
				},
			},
		}),
		Entry("outbound: invalid address", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Outbound[0].Address = "asdf"
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.outbound[0].address`,
						Message: `address has to be valid IP address`,
					},
				},
			},
		}),
		Entry("invalid networking.address", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Address = "asdf"
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.address`,
						Message: `address has to be valid IP address`,
					},
				},
			},
		}),
		Entry("multiple errors", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec = mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Interface: "127.0.0.1:8080:abc",
								Tags: map[string]string{
									"version": "",
								},
							},
						},
						Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
							{
								Interface: ":invalid",
								Service:   "",
							},
						},
					},
				}
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   "networking.inbound[0].interface",
						Message: "invalid format: expected format is DATAPLANE_IP:DATAPLANE_PORT:WORKLOAD_PORT , e.g. 192.168.0.100:9090:8080 or [2001:db8::1]:7070:6060",
					},
					{
						Field:   "networking.inbound[0].tags[\"service\"]",
						Message: "tag has to exist",
					},
					{
						Field:   "networking.inbound[0].tags[\"version\"]",
						Message: "tag value cannot be empty",
					},
					{
						Field:   "networking.outbound[0].interface",
						Message: "invalid format: expected format is DATAPLANE_IP:DATAPLANE_PORT where DATAPLANE_IP is optional. E.g. 127.0.0.1:9090, :9090, [::1]:8080",
					},
					{
						Field:   "networking.outbound[0].service",
						Message: "cannot be empty",
					},
				},
			},
		}),
		Entry("legacy - invalid inbound interface", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound[0].Interface = "asdf"
				validDataplane.Spec.Networking.Address = ""
				validDataplane.Spec.Networking.Inbound[0].Port = 0
				validDataplane.Spec.Networking.Inbound[0].ServicePort = 0
				validDataplane.Spec.Networking.Inbound[0].Address = ""
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   "networking.inbound[0].interface",
						Message: `invalid format: expected format is DATAPLANE_IP:DATAPLANE_PORT:WORKLOAD_PORT , e.g. 192.168.0.100:9090:8080 or [2001:db8::1]:7070:6060`,
					},
				},
			},
		}),
		Entry("legacy - mixing networking.address with legacy inbounds", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Address = "192.168.0.1"
				validDataplane.Spec.Networking.Inbound[0].Interface = "127.0.0.1:1234:5678"
				validDataplane.Spec.Networking.Inbound[0].Port = 0
				validDataplane.Spec.Networking.Inbound[0].ServicePort = 0
				validDataplane.Spec.Networking.Inbound[0].Address = ""
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.inbound[0].interface`,
						Message: `interface cannot be defined with networking.address. Replace it with port, servicePort and networking.address`,
					},
				},
			},
		}),
		Entry("legacy - mixing legacy inbounds config with new config", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound[0].Interface = "127.0.0.1:1234:5678"
				validDataplane.Spec.Networking.Inbound[0].Port = 1234
				validDataplane.Spec.Networking.Inbound[0].ServicePort = 5678
				validDataplane.Spec.Networking.Inbound[0].Address = "127.0.0.1"
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.inbound[0].interface`,
						Message: `interface cannot be defined with port. Replace it with port, servicePort and networking.address`,
					},
					{
						Field:   `networking.inbound[0].interface`,
						Message: `interface cannot be defined with servicePort. Replace it with port, servicePort and networking.address`,
					},
					{
						Field:   `networking.inbound[0].interface`,
						Message: `interface cannot be defined with address. Replace it with port, servicePort and networking.address`,
					},
					{
						Field:   "networking.inbound[0].interface",
						Message: "interface cannot be defined with networking.address. Replace it with port, servicePort and networking.address",
					},
				},
			},
		}),
		Entry("legacy - mixing legacy outbound config with new config", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Outbound[0].Interface = ":5678"
				validDataplane.Spec.Networking.Outbound[0].Port = 5678
				validDataplane.Spec.Networking.Outbound[0].Address = "127.0.0.1"
				return validDataplane
			},
			validationResult: &validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   `networking.outbound[0].interface`,
						Message: `interface cannot be defined with port. Replace it with port and address`,
					},
					{
						Field:   `networking.outbound[0].interface`,
						Message: `interface cannot be defined with address. Replace it with port and address`,
					},
				},
			},
		}),
	)

})
