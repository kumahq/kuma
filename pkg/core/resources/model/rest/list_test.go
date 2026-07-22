package rest_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest/v1alpha1"
	policies_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var _ = Describe("Unmarshal ResourceList", func() {
	Describe("old unversioned resources", func() {
		It("it should be possible to unmarshal JSON response from Kuma API Server", func() {
			// given
			content := `
			{
				"items": [
				 {
					"type": "Mesh",
					"name": "one",
					"mtls": {
					  "enabledBackend": "ca-1"
					}
				 },
				 {
					"type": "Mesh",
					"name": "two",
					"mtls": {
					  "enabledBackend": "ca-2"
					}
				 }
				],
				"next": "http://localhost:5681/meshes?offset=1"
			}`

			// when
			rsr := &rest.ResourceListReceiver{
				NewResource: func() core_model.Resource {
					return mesh.NewMeshResource()
				},
			}
			err := json.Unmarshal([]byte(content), rsr)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			rs := rsr.ResourceList
			// then
			Expect(rs.Items).To(HaveLen(2))
			Expect(rs.Items[0].GetMeta()).To(Equal(v1alpha1.ResourceMeta{
				Type: "Mesh",
				Name: "one",
			}))
			Expect(rs.Items[0].GetSpec()).To(matchers.MatchProto(&mesh_proto.Mesh{
				Mtls: &mesh_proto.Mesh_Mtls{
					EnabledBackend: "ca-1",
				},
			}))
			Expect(rs.Items[1].GetMeta()).To(Equal(v1alpha1.ResourceMeta{
				Type: "Mesh",
				Name: "two",
			}))
			Expect(rs.Items[1].GetSpec()).To(matchers.MatchProto(&mesh_proto.Mesh{
				Mtls: &mesh_proto.Mesh_Mtls{
					EnabledBackend: "ca-2",
				},
			}))
			Expect(*rs.Next).To(Equal("http://localhost:5681/meshes?offset=1"))
		})
	})

	Describe("new plugin originated policies", func() {
		It("it should be possible to unmarshal JSON response from Kuma API Server", func() {
			// given
			content := `
{
    "items": [
        {
            "type": "MeshTrafficPermission",
            "mesh": "default",
            "name": "mtp1",
            "spec": {
                "targetRef": {
                    "kind": "MeshService",
                    "name": "backend"
                },
                "from": [
                    {
                        "targetRef": {
                            "kind": "Mesh"
                        },
                        "default": {
                            "action": "Allow"
                        }
                    },
                    {
                        "targetRef": {
                            "kind": "MeshSubset",
                            "tags": {
                                "kuma.io/zone": "us-east"
                            }
                        },
                        "default": {
                            "action": "Deny"
                        }
                    },
                    {
                        "targetRef": {
                            "kind": "MeshService",
                            "name": "backend"
                        },
                        "default": {
                            "action": "AllowWithShadowDeny"
                        }
                    },
                    {
                        "targetRef": {
                            "kind": "MeshServiceSubset",
                            "name": "backend",
                            "tags": {
                                "version": "v1"
                            }
                        },
                        "default": {
                            "action": "Deny"
                        }
                    }
                ]
            }
        },
        {
            "type": "MeshTrafficPermission",
            "mesh": "default",
            "name": "mtp2",
            "spec": {
                "targetRef": {
                    "kind": "Mesh"
                },
                "from": [
                    {
                        "targetRef": {
                            "kind": "MeshSubset",
                            "tags": {
                                "kuma.io/zone": "us-east"
                            }
                        },
                        "default": {
                            "action": "Deny"
                        }
                    }
                ]
            }
        }
	],
	"next": "http://localhost:5681/meshes/default/meshtrafficpermissions?offset=1"
}`

			// when
			rsr := &rest.ResourceListReceiver{
				NewResource: func() core_model.Resource {
					return policies_api.NewMeshTrafficPermissionResource()
				},
			}
			err := json.Unmarshal([]byte(content), rsr)

			// then
			Expect(err).ToNot(HaveOccurred())

			rs := rsr.ResourceList

			Expect(rs.Items).To(HaveLen(2))

			Expect(rs.Items[0].GetMeta()).To(Equal(v1alpha1.ResourceMeta{
				Type: "MeshTrafficPermission",
				Mesh: "default",
				Name: "mtp1",
			}))
			Expect(rs.Items[0].GetSpec()).To(Equal(&policies_api.MeshTrafficPermission{
				TargetRef: &common_api.TargetRef{Kind: "MeshService", Name: pointer.To("backend")},
				From: &[]policies_api.From{
					{
						TargetRef: common_api.TargetRef{Kind: "Mesh"},
						Default: policies_api.Conf{
							Action: pointer.To[policies_api.Action]("Allow"),
						},
					},
					{
						TargetRef: common_api.TargetRef{Kind: "MeshSubset", Tags: &map[string]string{"kuma.io/zone": "us-east"}},
						Default: policies_api.Conf{
							Action: pointer.To[policies_api.Action]("Deny"),
						},
					},
					{
						TargetRef: common_api.TargetRef{Kind: "MeshService", Name: pointer.To("backend")},
						Default: policies_api.Conf{
							Action: pointer.To[policies_api.Action]("AllowWithShadowDeny"),
						},
					},
					{
						TargetRef: common_api.TargetRef{Kind: "MeshServiceSubset", Name: pointer.To("backend"), Tags: &map[string]string{"version": "v1"}},
						Default: policies_api.Conf{
							Action: pointer.To[policies_api.Action]("Deny"),
						},
					},
				},
			}))
			Expect(rs.Items[1].GetMeta()).To(Equal(v1alpha1.ResourceMeta{
				Type: "MeshTrafficPermission",
				Mesh: "default",
				Name: "mtp2",
			}))
			Expect(rs.Items[1].GetSpec()).To(Equal(&policies_api.MeshTrafficPermission{
				TargetRef: &common_api.TargetRef{Kind: "Mesh"},
				From: &[]policies_api.From{
					{
						TargetRef: common_api.TargetRef{Kind: "MeshSubset", Tags: &map[string]string{"kuma.io/zone": "us-east"}},
						Default: policies_api.Conf{
							Action: pointer.To[policies_api.Action]("Deny"),
						},
					},
				},
			}))
			Expect(*rs.Next).To(Equal("http://localhost:5681/meshes/default/meshtrafficpermissions?offset=1"))
		})
	})
})
