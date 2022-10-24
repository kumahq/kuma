package model_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshaccesslog_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshtrafficpermissions_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
)

var _ = Describe("Resource Utils", func() {

	Describe("ToJSON", func() {
		It("should marshal empty slice to '[]'", func() {
			// given
			var spec core_model.ResourceSpec = &meshaccesslog_proto.MeshAccessLog{
				TargetRef: &common_api.TargetRef{Kind: "Mesh"},
				From: []*meshaccesslog_proto.MeshAccessLog_From{{
					Default: &meshaccesslog_proto.MeshAccessLog_Conf{
						Backends: make([]*meshaccesslog_proto.MeshAccessLog_Backend, 0),
					},
				}},
			}
			// when
			bytes, err := core_model.ToJSON(meshaccesslog_proto.MeshAccessLogResourceTypeDescriptor, spec)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(MatchJSON(`
{
  "targetRef": {
    "kind": "Mesh"
  },
  "from": [
    {
      "default": {
        "backends": []
      }
    }
  ],
  "to": null
}
`))
		})
	})

	Describe("IsEmpty", func() {

		It("should return true if ResourceSpec is empty", func() {
			// given
			var spec core_model.ResourceSpec = &meshtrafficpermissions_proto.MeshTrafficPermission{}
			// when
			isEmpty := core_model.IsEmpty(spec)
			// then
			Expect(isEmpty).To(BeTrue())
		})

		It("should return false if ResourceSpec is not empty", func() {
			// given
			var spec core_model.ResourceSpec = &meshtrafficpermissions_proto.MeshTrafficPermission{
				TargetRef: &common_api.TargetRef{Kind: "Mesh"},
			}
			// when
			isEmpty := core_model.IsEmpty(spec)
			// then
			Expect(isEmpty).To(BeFalse())
		})
	})

	Describe("FullName", func() {

		It("should return joined package path and type name", func() {
			// given
			var spec core_model.ResourceSpec = &meshtrafficpermissions_proto.MeshTrafficPermission{}
			// when
			name := core_model.FullName(spec)
			// then
			Expect(name).To(Equal("github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1/MeshTrafficPermission"))
		})

	})

	Describe("Equal", func() {

		It("should return true if specs are equal", func() {
			// given
			var spec1 core_model.ResourceSpec = &meshtrafficpermissions_proto.MeshTrafficPermission{
				TargetRef: &common_api.TargetRef{Kind: "Mesh"},
				From: []*meshtrafficpermissions_proto.MeshTrafficPermission_From{
					{
						TargetRef: &common_api.TargetRef{
							Kind: "MeshService",
							Name: "backend",
						},
						Default: &meshtrafficpermissions_proto.MeshTrafficPermission_Conf{Action: "ALLOW"},
					},
					{
						TargetRef: &common_api.TargetRef{
							Kind: "MeshService",
							Name: "web",
						},
						Default: &meshtrafficpermissions_proto.MeshTrafficPermission_Conf{Action: "DENY"},
					},
					{
						TargetRef: &common_api.TargetRef{
							Kind: "MeshSubset",
							Tags: map[string]string{
								"version": "v3",
							},
						},
						Default: &meshtrafficpermissions_proto.MeshTrafficPermission_Conf{Action: "ALLOW"},
					},
				},
			}
			var spec2 core_model.ResourceSpec = &meshtrafficpermissions_proto.MeshTrafficPermission{
				TargetRef: &common_api.TargetRef{Kind: "Mesh"},
				From: []*meshtrafficpermissions_proto.MeshTrafficPermission_From{
					{
						TargetRef: &common_api.TargetRef{
							Kind: "MeshService",
							Name: "backend",
						},
						Default: &meshtrafficpermissions_proto.MeshTrafficPermission_Conf{Action: "ALLOW"},
					},
					{
						TargetRef: &common_api.TargetRef{
							Kind: "MeshService",
							Name: "web",
						},
						Default: &meshtrafficpermissions_proto.MeshTrafficPermission_Conf{Action: "DENY"},
					},
					{
						TargetRef: &common_api.TargetRef{
							Kind: "MeshSubset",
							Tags: map[string]string{
								"version": "v3",
							},
						},
						Default: &meshtrafficpermissions_proto.MeshTrafficPermission_Conf{Action: "ALLOW"},
					},
				},
			}
			// when
			equal := core_model.Equal(spec1, spec2)
			// then
			Expect(equal).To(BeTrue())
		})

		It("should return false if specs are different", func() {
			// given
			var spec1 core_model.ResourceSpec = &meshtrafficpermissions_proto.MeshTrafficPermission{
				TargetRef: &common_api.TargetRef{Kind: "Mesh"},
				From: []*meshtrafficpermissions_proto.MeshTrafficPermission_From{
					{
						TargetRef: &common_api.TargetRef{
							Kind: "MeshService",
							Name: "backend",
						},
						Default: &meshtrafficpermissions_proto.MeshTrafficPermission_Conf{Action: "ALLOW"},
					},
					{
						TargetRef: &common_api.TargetRef{
							Kind: "MeshService",
							Name: "web",
						},
						Default: &meshtrafficpermissions_proto.MeshTrafficPermission_Conf{Action: "DENY"},
					},
					{
						TargetRef: &common_api.TargetRef{
							Kind: "MeshSubset",
							Tags: map[string]string{
								"version": "v3",
							},
						},
						Default: &meshtrafficpermissions_proto.MeshTrafficPermission_Conf{Action: "ALLOW"},
					},
				},
			}
			var spec2 core_model.ResourceSpec = &meshtrafficpermissions_proto.MeshTrafficPermission{
				TargetRef: &common_api.TargetRef{Kind: "Mesh"},
				From: []*meshtrafficpermissions_proto.MeshTrafficPermission_From{
					{
						TargetRef: &common_api.TargetRef{
							Kind: "MeshService",
							Name: "backend",
						},
						Default: &meshtrafficpermissions_proto.MeshTrafficPermission_Conf{Action: "ALLOW"},
					},
					{
						TargetRef: &common_api.TargetRef{
							Kind: "MeshService",
							Name: "web",
						},
						Default: &meshtrafficpermissions_proto.MeshTrafficPermission_Conf{Action: "DENY"},
					},
					{
						TargetRef: &common_api.TargetRef{
							Kind: "MeshSubset",
							Tags: map[string]string{
								"version": "v5", // different from 'v3'
							},
						},
						Default: &meshtrafficpermissions_proto.MeshTrafficPermission_Conf{Action: "ALLOW"},
					},
				},
			}
			// when
			equal := core_model.Equal(spec1, spec2)
			// then
			Expect(equal).To(BeFalse())
		})
	})
})
