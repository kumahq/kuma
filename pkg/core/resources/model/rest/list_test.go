package rest_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Unmarshal ResourceList", func() {
	Describe("old unversioned resources", func() {
		It("it should be possible to unmarshal JSON response from Kuma API Server", func() {
			// given
			content := `
			{
				"items": [
				 {
					"type": "TrafficRoute",
					"mesh": "default",
					"name": "one",
					"conf": {
					  "destination": {
						"path": "/example"
					  }
					}
				 },
				 {
					"type": "TrafficRoute",
					"mesh": "demo",
					"name": "two",
					"conf": {
					  "destination": {
						"path": "/another"
					  }
					}
				 }
				],
				"next": "http://localhost:5681/meshes/default/traffic-routes?offset=1"
			}`

			// when
			rsr := &rest.ResourceListReceiver{
				NewResource: func() core_model.Resource {
					return mesh.NewTrafficRouteResource()
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
				Type: "TrafficRoute",
				Mesh: "default",
				Name: "one",
			}))
			Expect(rs.Items[0].GetSpec()).To(matchers.MatchProto(&mesh_proto.TrafficRoute{
				Conf: &mesh_proto.TrafficRoute_Conf{
					Destination: map[string]string{
						"path": "/example",
					},
				},
			}))
			Expect(rs.Items[1].GetMeta()).To(Equal(v1alpha1.ResourceMeta{
				Type: "TrafficRoute",
				Mesh: "demo",
				Name: "two",
			}))
			Expect(rs.Items[1].GetSpec()).To(matchers.MatchProto(&mesh_proto.TrafficRoute{
				Conf: &mesh_proto.TrafficRoute_Conf{
					Destination: map[string]string{
						"path": "/another",
					},
				},
			}))
			Expect(*rs.Next).To(Equal("http://localhost:5681/meshes/default/traffic-routes?offset=1"))
		})
		It("it should be possible to unmarshal JSON response with protobuf field types", func() {
			// given
			content := `
{
	"items": [
	 {
		"type": "CircuitBreaker",
		"mesh": "default",
		"name": "one",
		"sources": [
			{
				"match": {
					"kuma.io/service": "*"
			  }
			}
		],
		"destinations": [
			{
				"match": {
					"kuma.io/service": "*"
				}
			}
		],
		"conf": {
			"interval": "99s",
			"detectors": {
				"totalErrors": {
					"consecutive": 3
				}
			}
		}
	},
	{
		"type": "CircuitBreaker",
		"mesh": "default",
		"name": "two",
		"sources": [
			{
				"match": {
					"kuma.io/service": "*"
			  }
			}
		],
		"destinations": [
			{
				"match": {
					"kuma.io/service": "*"
				}
			}
		],
		"conf": {
			"interval": "11s",
			"detectors": {
				"totalErrors": {
					"consecutive": 10
				}
			}
		}	
	}
	],
	"next": "http://localhost:5681/meshes/default/circuit-breakers?offset=1"
}`

			// when
			rsr := &rest.ResourceListReceiver{
				NewResource: func() core_model.Resource {
					return mesh.NewCircuitBreakerResource()
				},
			}
			err := json.Unmarshal([]byte(content), rsr)

			// then
			Expect(err).ToNot(HaveOccurred())

			rs := rsr.ResourceList

			Expect(rs.Items).To(HaveLen(2))
			Expect(rs.Items[0].GetMeta()).To(Equal(v1alpha1.ResourceMeta{
				Type: "CircuitBreaker",
				Mesh: "default",
				Name: "one",
			}))
			Expect(rs.Items[0].GetSpec()).To(matchers.MatchProto(&mesh_proto.CircuitBreaker{
				Sources: []*mesh_proto.Selector{
					{Match: map[string]string{"kuma.io/service": "*"}},
				},
				Destinations: []*mesh_proto.Selector{
					{Match: map[string]string{"kuma.io/service": "*"}},
				},
				Conf: &mesh_proto.CircuitBreaker_Conf{
					Interval: util_proto.Duration(99 * time.Second),
					Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
						TotalErrors: &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{
							Consecutive: util_proto.UInt32(3),
						},
					},
				},
			}))
			Expect(rs.Items[1].GetMeta()).To(Equal(v1alpha1.ResourceMeta{
				Type: "CircuitBreaker",
				Mesh: "default",
				Name: "two",
			}))
			Expect(rs.Items[1].GetSpec()).To(matchers.MatchProto(&mesh_proto.CircuitBreaker{
				Sources: []*mesh_proto.Selector{
					{Match: map[string]string{"kuma.io/service": "*"}},
				},
				Destinations: []*mesh_proto.Selector{
					{Match: map[string]string{"kuma.io/service": "*"}},
				},
				Conf: &mesh_proto.CircuitBreaker_Conf{
					Interval: util_proto.Duration(11 * time.Second),
					Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
						TotalErrors: &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{
							Consecutive: util_proto.UInt32(10),
						},
					},
				},
			}))
			Expect(*rs.Next).To(Equal("http://localhost:5681/meshes/default/circuit-breakers?offset=1"))
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
				TargetRef: &common_api.TargetRef{Kind: "MeshService", Name: "backend"},
				From: []policies_api.From{
					{
						TargetRef: common_api.TargetRef{Kind: "Mesh"},
						Default: policies_api.Conf{
							Action: "Allow",
						},
					},
					{
						TargetRef: common_api.TargetRef{Kind: "MeshSubset", Tags: map[string]string{"kuma.io/zone": "us-east"}},
						Default: policies_api.Conf{
							Action: "Deny",
						},
					},
					{
						TargetRef: common_api.TargetRef{Kind: "MeshService", Name: "backend"},
						Default: policies_api.Conf{
							Action: "AllowWithShadowDeny",
						},
					},
					{
						TargetRef: common_api.TargetRef{Kind: "MeshServiceSubset", Name: "backend", Tags: map[string]string{"version": "v1"}},
						Default: policies_api.Conf{
							Action: "Deny",
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
				From: []policies_api.From{
					{
						TargetRef: common_api.TargetRef{Kind: "MeshSubset", Tags: map[string]string{"kuma.io/zone": "us-east"}},
						Default: policies_api.Conf{
							Action: "Deny",
						},
					},
				},
			}))
			Expect(*rs.Next).To(Equal("http://localhost:5681/meshes/default/meshtrafficpermissions?offset=1"))
		})
	})
})
