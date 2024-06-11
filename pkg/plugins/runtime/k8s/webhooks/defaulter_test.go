package webhooks_test

import (
	"context"

	jsonpatch "github.com/evanphx/json-patch/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	k8s_resources "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	. "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
)

var _ = Describe("Defaulter", func() {
	var converter k8s_common.Converter

	BeforeEach(func() {
		converter = k8s_resources.NewSimpleConverter()
	})

	type testCase struct {
		inputObject string
		expected    string
		kind        string
		checker     ResourceAdmissionChecker
	}

	allowedUsers := []string{"system:serviceaccount:kube-system:generic-garbage-collector", "system:serviceaccount:kuma-system:kuma-control-plane"}

	globalChecker := func() ResourceAdmissionChecker {
		return ResourceAdmissionChecker{
			AllowedUsers:                 allowedUsers,
			Mode:                         core.Global,
			FederatedZone:                false,
			DisableOriginLabelValidation: false,
		}
	}

	zoneChecker := func(federatedZone, originValidation bool) ResourceAdmissionChecker {
		return ResourceAdmissionChecker{
			AllowedUsers:                 allowedUsers,
			Mode:                         core.Zone,
			FederatedZone:                federatedZone,
			DisableOriginLabelValidation: !originValidation,
		}
	}

	DescribeTable("should apply defaults on a target object",
		func(given testCase) {
			// given
			handler := DefaultingWebhookFor(scheme, converter, given.checker)

			req := kube_admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: kube_types.UID("12345"),
					Object: kube_runtime.RawExtension{
						Raw: []byte(given.inputObject),
					},
					Kind: kube_meta.GroupVersionKind{
						Kind: given.kind,
					},
				},
			}

			// when
			resp := handler.Handle(context.Background(), req)

			// then
			Expect(resp.UID).To(Equal(kube_types.UID("12345")))
			Expect(resp.Result.Message).To(Equal(""))
			Expect(resp.Allowed).To(BeTrue())

			var actual []byte
			if len(resp.Patch) == 0 {
				actual = []byte(given.inputObject)
			} else {
				patch, err := jsonpatch.DecodePatch(resp.Patch)
				Expect(err).ToNot(HaveOccurred())
				actual, err = patch.Apply([]byte(given.inputObject))
				Expect(err).ToNot(HaveOccurred())
			}
			Expect(actual).To(MatchJSON(given.expected))
		},
		Entry("should apply defaults to empty conf", testCase{
			checker: globalChecker(),
			kind:    string(mesh.MeshType),
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "Mesh",
              "metadata": {
				"name": "empty",
				"creationTimestamp": null
              },
              "spec": {
				"metrics": {
				  "backends": [
					{
					  "type": "prometheus"
					}
				  ]
				}
              }
            }
`,
			expected: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "Mesh",
              "metadata": {
				"name": "empty",
<<<<<<< HEAD
				"creationTimestamp": null
=======
				"creationTimestamp": null,
				"annotations": {
				  "kuma.io/display-name": "empty"
				}
>>>>>>> da824ce57 (fix(kuma-cp): mistakenly setting 'kuma.io/display-name' as label (#10430))
              },
              "spec": {
				"metrics": {
				  "backends": [
					{
					  "type": "prometheus",
					  "conf": {
						"path": "/metrics",
						"port": 5670,
						"tags": {
						  "kuma.io/service": "dataplane-metrics"
						},
						"tls": {}
					  }
					}
				  ]
				}
              }
            }
`,
		}),
		Entry("should not override non-empty spec fields", testCase{
			checker: globalChecker(),
			kind:    string(mesh.MeshType),
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "Mesh",
              "metadata": {
				"name": "empty",
				"creationTimestamp": null
              },
              "spec": {
				"metrics": {
				  "backends": [
					{
					  "type": "prometheus",
					  "conf": {
						"path": "/dont/override"
					  }
					}
				  ]
				}
              }
            }
`,
			expected: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "Mesh",
              "metadata": {
				"name": "empty",
<<<<<<< HEAD
				"creationTimestamp": null
=======
				"creationTimestamp": null,
				"annotations": {
				  "kuma.io/display-name": "empty"
				}
>>>>>>> da824ce57 (fix(kuma-cp): mistakenly setting 'kuma.io/display-name' as label (#10430))
              },
              "spec": {
				"metrics": {
				  "backends": [
					{
					  "type": "prometheus",
					  "conf": {
						"path": "/dont/override",
						"port": 5670,
						"tags": {
						  "kuma.io/service": "dataplane-metrics"
						},
						"tls": {}
					  }
					}
				  ]
				}
              }
            }
`,
		}),
		Entry("should not override mesh label if it's already set", testCase{
			checker: globalChecker(),
			kind:    string(mesh.TrafficRouteType),
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "TrafficRoute",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null,
                "labels": {
                  "kuma.io/mesh": "my-mesh-1"
                }
              },
              "spec": {}
            }
`,
			expected: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "TrafficRoute",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null,
                "labels": {
<<<<<<< HEAD
                  "kuma.io/mesh": "my-mesh-1"
=======
                  "kuma.io/mesh": "my-mesh-1",
                  "k8s.kuma.io/namespace": "example"
                },
                "annotations": {
                  "kuma.io/display-name": "empty"
>>>>>>> da824ce57 (fix(kuma-cp): mistakenly setting 'kuma.io/display-name' as label (#10430))
                }
              },
              "spec": {}
            }
`,
		}),
		Entry("should set mesh label when apply new policy on Zone", testCase{
			checker: zoneChecker(true, true),
			kind:    string(v1alpha1.MeshTrafficPermissionType),
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "MeshTrafficPermission",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null,
                "labels": {
                  "kuma.io/origin": "zone"
                }
              },
              "spec": {
                "targetRef": {}
              }
            }
`,
			expected: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "MeshTrafficPermission",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null,
                "labels": {
                  "kuma.io/origin": "zone",
<<<<<<< HEAD
                  "kuma.io/mesh": "default"
=======
                  "kuma.io/mesh": "default",
                  "kuma.io/policy-role": "workload-owner",
                  "k8s.kuma.io/namespace": "example"
                },
                "annotations": {
                  "kuma.io/display-name": "empty"
>>>>>>> da824ce57 (fix(kuma-cp): mistakenly setting 'kuma.io/display-name' as label (#10430))
                }
              },
              "spec": {
                "targetRef": {}
              }
            }
`,
		}),
		Entry("should set mesh and origin label when origin validation is disabled, federated zone", testCase{
			checker: zoneChecker(true, false),
			kind:    string(v1alpha1.MeshTrafficPermissionType),
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "MeshTrafficPermission",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null
              },
              "spec": {
                "targetRef": {}
              }
            }
`,
			expected: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "MeshTrafficPermission",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null,
                "labels": {
<<<<<<< HEAD
                  "kuma.io/origin": "zone",
                  "kuma.io/mesh": "default"
=======
                  "k8s.kuma.io/namespace": "example",
                  "kuma.io/mesh": "default",
                  "kuma.io/origin": "zone",
                  "kuma.io/policy-role": "workload-owner"
                },
                "annotations": {
                  "kuma.io/display-name": "empty"
>>>>>>> da824ce57 (fix(kuma-cp): mistakenly setting 'kuma.io/display-name' as label (#10430))
                }
              },
              "spec": {
                "targetRef": {}
              }
            }
`,
		}),
		Entry("should set mesh and origin label when origin validation is disabled, non-federated zone", testCase{
			checker: zoneChecker(false, false),
			kind:    string(v1alpha1.MeshTrafficPermissionType),
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "MeshTrafficPermission",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null
              },
              "spec": {
                "targetRef": {}
              }
            }
`,
			expected: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "MeshTrafficPermission",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null,
                "labels": {
<<<<<<< HEAD
                  "kuma.io/origin": "zone",
                  "kuma.io/mesh": "default"
=======
                  "k8s.kuma.io/namespace": "example",
                  "kuma.io/mesh": "default",
                  "kuma.io/origin": "zone",
                  "kuma.io/policy-role": "workload-owner"
                },
                "annotations": {
                  "kuma.io/display-name": "empty"
>>>>>>> da824ce57 (fix(kuma-cp): mistakenly setting 'kuma.io/display-name' as label (#10430))
                }
              },
              "spec": {
                "targetRef": {}
              }
            }
`,
		}),
		Entry("should set mesh and origin label on DPP", testCase{
			checker: zoneChecker(true, true),
			kind:    string(mesh.DataplaneType),
			inputObject: `
            {
              "apiVersion":"kuma.io/v1alpha1",
              "kind":"Dataplane",
              "mesh":"demo",
              "metadata":{
                "namespace":"example",
                "name":"empty",
                "creationTimestamp":null
              },
              "spec":{
                "networking": {
                  "address": "127.0.0.1",
                  "inbound": [
                    {
                      "port": 11011,
                      "tags": {
                        "kuma.io/service": "backend"
                      }
                    }
                  ]
                }
              }
            }`,
			expected: `
            {
              "apiVersion":"kuma.io/v1alpha1",
              "kind":"Dataplane",
              "mesh":"demo",
              "metadata":{
                "namespace":"example",
                "name":"empty",
                "creationTimestamp":null,
                "labels": {
<<<<<<< HEAD
                  "kuma.io/origin": "zone",
                  "kuma.io/mesh": "default"
=======
                  "k8s.kuma.io/namespace": "example",
                  "kuma.io/mesh": "default",
                  "kuma.io/origin": "zone"
                },
                "annotations": {
                  "kuma.io/display-name": "empty"
>>>>>>> da824ce57 (fix(kuma-cp): mistakenly setting 'kuma.io/display-name' as label (#10430))
                }
              },
              "spec":{
                "networking": {
                  "address": "127.0.0.1",
                  "inbound": [
                    {
                      "port": 11011,
                      "tags": {
                        "kuma.io/service": "backend"
                      }
                    }
                  ]
                }
              }
            }`,
		}),
		Entry("should not add origin label on Global", testCase{
			checker: globalChecker(),
			kind:    string(v1alpha1.MeshTrafficPermissionType),
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "MeshTrafficPermission",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null
              },
              "spec": {
                "targetRef": {}
              }
            }
`,
			expected: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "MeshTrafficPermission",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null,
                "labels": {
<<<<<<< HEAD
                  "kuma.io/mesh": "default"
=======
                  "k8s.kuma.io/namespace": "example",
                  "kuma.io/mesh": "default",
                  "kuma.io/policy-role": "workload-owner"
                },
                "annotations": {
                  "kuma.io/display-name": "empty"
>>>>>>> da824ce57 (fix(kuma-cp): mistakenly setting 'kuma.io/display-name' as label (#10430))
                }
              },
              "spec": {
                "targetRef": {}
              }
            }
`,
		}),
	)
})
