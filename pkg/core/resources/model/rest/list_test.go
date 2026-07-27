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
                "rules": [
                    {
                        "default": {
                            "allow": [
                                {
                                    "spiffeID": {
                                        "type": "Exact",
                                        "value": "spiffe://mesh/sa/allowed"
                                    }
                                }
                            ]
                        }
                    },
                    {
                        "default": {
                            "deny": [
                                {
                                    "spiffeID": {
                                        "type": "Prefix",
                                        "value": "spiffe://zone-us-east"
                                    }
                                }
                            ]
                        }
                    },
                    {
                        "default": {
                            "allowWithShadowDeny": [
                                {
                                    "spiffeID": {
                                        "type": "Exact",
                                        "value": "spiffe://mesh/sa/backend"
                                    }
                                }
                            ]
                        }
                    },
                    {
                        "default": {
                            "deny": [
                                {
                                    "sni": {
                                        "type": "Exact",
                                        "value": "backend.svc"
                                    }
                                }
                            ]
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
                "rules": [
                    {
                        "default": {
                            "deny": [
                                {
                                    "spiffeID": {
                                        "type": "Prefix",
                                        "value": "spiffe://zone-us-east"
                                    }
                                }
                            ]
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
				Rules: &[]policies_api.Rule{
					{
						Default: policies_api.RuleConf{
							Allow: &[]common_api.Match{
								{SpiffeID: &common_api.SpiffeIDMatch{Type: common_api.ExactMatchType, Value: "spiffe://mesh/sa/allowed"}},
							},
						},
					},
					{
						Default: policies_api.RuleConf{
							Deny: &[]common_api.Match{
								{SpiffeID: &common_api.SpiffeIDMatch{Type: common_api.PrefixMatchType, Value: "spiffe://zone-us-east"}},
							},
						},
					},
					{
						Default: policies_api.RuleConf{
							AllowWithShadowDeny: &[]common_api.Match{
								{SpiffeID: &common_api.SpiffeIDMatch{Type: common_api.ExactMatchType, Value: "spiffe://mesh/sa/backend"}},
							},
						},
					},
					{
						Default: policies_api.RuleConf{
							Deny: &[]common_api.Match{
								{SNI: &common_api.SNIMatch{Type: common_api.SNIExactMatchType, Value: "backend.svc"}},
							},
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
				Rules: &[]policies_api.Rule{
					{
						Default: policies_api.RuleConf{
							Deny: &[]common_api.Match{
								{SpiffeID: &common_api.SpiffeIDMatch{Type: common_api.PrefixMatchType, Value: "spiffe://zone-us-east"}},
							},
						},
					},
				},
			}))
			Expect(*rs.Next).To(Equal("http://localhost:5681/meshes/default/meshtrafficpermissions?offset=1"))
		})
	})
})
