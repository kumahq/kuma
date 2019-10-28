package mesh_test

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/validators"
	"github.com/Kong/kuma/pkg/core/validators/apis/mesh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dataplane", func() {

	var validDataplane core_mesh.DataplaneResource

	BeforeEach(func() {
		validDataplane = core_mesh.DataplaneResource{
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
	})

	It("should pass validation", func() {
		// when
		err := mesh.ValidateDataplane(&validDataplane)

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	type testCase struct {
		dataplane func() core_mesh.DataplaneResource
		err       string
	}
	DescribeTable("should catch validation errors",
		func(given testCase) {
			// when
			dp := given.dataplane()
			err := mesh.ValidateDataplane(&dp)

			// then
			Expect(err).To(HaveOccurred())
			Expect(validators.IsValidationError(err)).To(BeTrue())
			Expect(err).To(MatchError(given.err))
		},
		Entry("empty inbound interface", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound[0].Interface = ""
				return validDataplane
			},
			err: `validation error: Inbound[0]: Interface: invalid format: expected ^(?P<dataplane_ip>(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)):(?P<dataplane_port>[0-9]{1,5}):(?P<workload_port>[0-9]{1,5})$, got ""`,
		}),
		Entry("invalid inbound interface", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound[0].Interface = "asdf"
				return validDataplane
			},
			err: `validation error: Inbound[0]: Interface: invalid format: expected ^(?P<dataplane_ip>(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)):(?P<dataplane_port>[0-9]{1,5}):(?P<workload_port>[0-9]{1,5})$, got "asdf"`,
		}),
		Entry("inbound: empty service tag", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound[0].Tags[mesh_proto.ServiceTag] = ""
				return validDataplane
			},
			err: `validation error: Inbound[0]: "service" tag has to exist and be non empty`,
		}),
		Entry("inbound: empty service tag", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Inbound[0].Tags = map[string]string{}
				return validDataplane
			},
			err: `validation error: Inbound[0]: "service" tag has to exist and be non empty`,
		}),
		Entry("outbound: empty service tag", testCase{
			dataplane: func() core_mesh.DataplaneResource {
				validDataplane.Spec.Networking.Outbound[0].Service = ""
				return validDataplane
			},
			err: `validation error: Outbound[0]: Service cannot be empty`,
		}),
	)

})
