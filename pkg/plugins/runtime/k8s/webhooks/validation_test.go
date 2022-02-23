package webhooks_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_resources "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
)

var _ = Describe("Validation", func() {

	var converter k8s_common.Converter

	BeforeEach(func() {
		converter = k8s_resources.NewSimpleConverter()
	})

	type testCase struct {
		objTemplate core_model.ResourceSpec
		obj         string
		mode        core.CpMode
		resp        kube_admission.Response
		username    string
		operation   admissionv1.Operation
	}
	DescribeTable("Validation",
		func(given testCase) {
			// given
			webhook := &kube_admission.Webhook{
				Handler: webhooks.NewValidatingWebhook(converter, core_registry.Global(), k8s_registry.Global(), given.mode, "system:serviceaccount:kuma-system:kuma-control-plane"),
			}
			Expect(webhook.InjectScheme(scheme)).To(Succeed())

			obj, err := k8s_registry.Global().NewObject(given.objTemplate)
			Expect(err).ToNot(HaveOccurred())
			req := kube_admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: kube_types.UID("12345"),
					Object: kube_runtime.RawExtension{
						Raw: []byte(given.obj),
					},
					Kind: kube_meta.GroupVersionKind{
						Group:   obj.GetObjectKind().GroupVersionKind().Group,
						Version: obj.GetObjectKind().GroupVersionKind().Version,
						Kind:    obj.GetObjectKind().GroupVersionKind().Kind,
					},
					UserInfo: authenticationv1.UserInfo{
						Username: given.username,
					},
					Operation: given.operation,
				},
			}

			// when
			resp := webhook.Handle(context.Background(), req)

			// then
			Expect(resp).To(Equal(given.resp))
		},
		Entry("should pass validation", testCase{
			mode:        core.Standalone,
			objTemplate: &mesh_proto.TrafficRoute{},
			username:    "cli-user",
			obj: `
            {
              "apiVersion":"kuma.io/v1alpha1",
              "kind":"TrafficRoute",
              "mesh":"demo",
              "metadata":{
                "name":"empty",
                "creationTimestamp":null
              },
              "spec":{
                "sources":[
                  {
                    "match":{
                      "kuma.io/service":"web"
                    }
                  }
                ],
                "destinations":[
                  {
                    "match":{
                      "kuma.io/service":"backend"
                    }
                  }
                ],
                "conf":{
                 "split":[
                  {
                    "weight":100,
                    "destination":{
                      "kuma.io/service":"backend"
                    }
                  }
                ]
                }
              }
            }`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					UID:     "12345",
					Allowed: true,
					Result: &kube_meta.Status{
						Code: 200,
					},
				},
			},
			operation: admissionv1.Create,
		}),
		Entry("should pass validation for synced policy from Global to Zone", testCase{
			mode:        core.Zone,
			objTemplate: &mesh_proto.TrafficRoute{},
			username:    "system:serviceaccount:kuma-system:kuma-control-plane",
			obj: `
            {
              "apiVersion":"kuma.io/v1alpha1",
              "kind":"TrafficRoute",
              "mesh":"demo",
              "metadata":{
                "name":"empty",
                "creationTimestamp":null,
                "annotations": {
                  "k8s.kuma.io/synced": "true"
                }
              },
              "spec":{
                "sources":[
                  {
                    "match":{
                      "kuma.io/service":"web"
                    }
                  }
                ],
                "destinations":[
                  {
                    "match":{
                      "kuma.io/service":"backend"
                    }
                  }
                ],
                "conf":{
                 "split":[
                  {
                    "weight":100,
                    "destination":{
                      "kuma.io/service":"backend"
                    }
                  }
                ]
                }
              }
            }`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					UID:     "12345",
					Allowed: true,
					Result: &kube_meta.Status{
						Code: 200,
					},
				},
			},
			operation: admissionv1.Create,
		}),
		Entry("should pass validation for synced policy from Zone to Global", testCase{
			mode:        core.Zone,
			objTemplate: &mesh_proto.Dataplane{},
			username:    "system:serviceaccount:kuma-system:kuma-control-plane",
			obj: `
            {
              "apiVersion":"kuma.io/v1alpha1",
              "kind":"Dataplane",
              "mesh":"demo",
              "metadata":{
                "namespace":"example",
                "name":"empty",
                "creationTimestamp":null,
                "annotations": {
                  "k8s.kuma.io/synced": "true"
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
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					UID:     "12345",
					Allowed: true,
					Result: &kube_meta.Status{
						Code: 200,
					},
				},
			},
			operation: admissionv1.Create,
		}),
		Entry("should pass validation for not synced Dataplane in Zone", testCase{
			mode:        core.Zone,
			objTemplate: &mesh_proto.Dataplane{},
			username:    "cli-user",
			obj: `
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
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					UID:     "12345",
					Allowed: true,
					Result: &kube_meta.Status{
						Code: 200,
					},
				},
			},
			operation: admissionv1.Create,
		}),
		Entry("should fail validation due to invalid spec", testCase{
			mode:        core.Global,
			objTemplate: &mesh_proto.TrafficRoute{},
			username:    "cli-user",
			obj: `
			{
			  "apiVersion": "kuma.io/v1alpha1",
			  "kind": "TrafficRoute",
			  "mesh": "demo",
			  "metadata": {
				"name": "empty",
				"creationTimestamp": null
			  },
			  "spec": {
			  }
			}
			`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					UID:     "12345",
					Allowed: false,
					Result: &kube_meta.Status{
						Status:  "Failure",
						Message: "spec.sources: must have at least one element; spec.destinations: must have at least one element; spec.conf: cannot be empty",
						Reason:  "Invalid",
						Details: &kube_meta.StatusDetails{
							Name: "empty",
							Kind: "TrafficRoute",
							Causes: []kube_meta.StatusCause{
								{
									Type:    "FieldValueInvalid",
									Message: "must have at least one element",
									Field:   "spec.sources",
								},
								{
									Type:    "FieldValueInvalid",
									Message: "must have at least one element",
									Field:   "spec.destinations",
								},
								{
									Type:    "FieldValueInvalid",
									Message: "cannot be empty",
									Field:   "spec.conf",
								},
							},
						},
						Code: 422,
					},
				},
			},
			operation: admissionv1.Create,
		}),
		Entry("should fail validation due to applying policy manually on Zone CP", testCase{
			mode:        core.Zone,
			objTemplate: &mesh_proto.TrafficRoute{},
			username:    "cli-user",
			obj: `
			{
			  "apiVersion": "kuma.io/v1alpha1",
			  "kind": "TrafficRoute",
			  "mesh": "demo",
			  "metadata": {
				"name": "empty",
				"creationTimestamp": null
			  }
			}
			`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					UID:     "12345",
					Allowed: false,
					Result: &kube_meta.Status{
						Status:  "Failure",
						Message: "Operation not allowed. Kuma resources like TrafficRoute can be updated or deleted only from the GLOBAL control plane and not from a ZONE control plane.",
						Reason:  "Forbidden",
						Details: &kube_meta.StatusDetails{
							Causes: []kube_meta.StatusCause{
								{
									Type:    "FieldValueInvalid",
									Message: "cannot be empty",
									Field:   "metadata.annotations[kuma.io/synced]",
								},
							},
						},
						Code: 403,
					},
				},
			},
			operation: admissionv1.Create,
		}),
		Entry("should fail validation due to applying Dataplane manually on Global CP", testCase{
			mode:        core.Global,
			objTemplate: &mesh_proto.Dataplane{},
			username:    "cli-user",
			obj: `
			{
			  "apiVersion": "kuma.io/v1alpha1",
			  "kind": "Dataplane",
			  "mesh": "demo",
			  "metadata": {
				"namespace": "example",
				"name": "empty",
				"creationTimestamp": null
			  }
			}
			`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					UID:     "12345",
					Allowed: false,
					Result: &kube_meta.Status{
						Status:  "Failure",
						Message: "Operation not allowed. Kuma resources like Dataplane can be updated or deleted only from the ZONE control plane and not from a GLOBAL control plane.",
						Reason:  "Forbidden",
						Details: &kube_meta.StatusDetails{
							Causes: []kube_meta.StatusCause{
								{
									Type:    "FieldValueInvalid",
									Message: "cannot be empty",
									Field:   "metadata.annotations[kuma.io/synced]",
								},
							},
						},
						Code: 403,
					},
				},
			},
			operation: admissionv1.Create,
		}),
		Entry("should pass validation due to applying Zone on Global CP", testCase{
			mode:        core.Global,
			objTemplate: &system_proto.Zone{},
			username:    "cli-user",
			obj: `
			{
			  "apiVersion": "kuma.io/v1alpha1",
			  "kind": "Zone",
			  "mesh": "default",
			  "metadata": {
				"name": "zone-1",
				"creationTimestamp": null
			  },
			  "spec": {
			    "ingress": {
			      "address": "192.168.0.1:10001"
			    }
			  }			
			}
			`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					UID:     "12345",
					Allowed: true,
					Result: &kube_meta.Status{
						Code: 200,
					},
				},
			},
			operation: admissionv1.Create,
		}),
		Entry("should fail validation due to applying Zone on Zone CP", testCase{
			mode:        core.Zone,
			objTemplate: &system_proto.Zone{},
			username:    "cli-user",
			obj: `
			{
			  "apiVersion": "kuma.io/v1alpha1",
			  "kind": "Zone",
			  "mesh": "default",
			  "metadata": {
				"name": "zone-1",
				"creationTimestamp": null
			  },
			  "spec": {
			    "ingress": {
			      "address": "192.168.0.1:10001"
			    }
			  }			
			}
			`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					UID:     "12345",
					Allowed: false,
					Result: &kube_meta.Status{
						Status:  "Failure",
						Message: "Zone resource can only be applied on CP with mode: [global]",
						Reason:  "Forbidden",
						Code:    403,
					},
				},
			},
			operation: admissionv1.Create,
		}),
		Entry("should fail validation due to applying Zone on Standalone CP", testCase{
			mode:        core.Standalone,
			objTemplate: &system_proto.Zone{},
			username:    "cli-user",
			obj: `
			{
			  "apiVersion": "kuma.io/v1alpha1",
			  "kind": "Zone",
			  "mesh": "default",
			  "metadata": {
				"name": "zone-1",
				"creationTimestamp": null
			  },
			  "spec": {
			    "ingress": {
			      "address": "192.168.0.1:10001"
			    }
			  }			
			}
			`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					UID:     "12345",
					Allowed: false,
					Result: &kube_meta.Status{
						Status:  "Failure",
						Message: "Zone resource can only be applied on CP with mode: [global]",
						Reason:  "Forbidden",
						Code:    403,
					},
				},
			},
			operation: admissionv1.Create,
		}),
		Entry("should fail validation on missing mesh object", testCase{
			mode:        core.Zone,
			objTemplate: &mesh_proto.TrafficRoute{},
			username:    "system:serviceaccount:kuma-system:kuma-control-plane",
			obj: `
            {
              "apiVersion":"kuma.io/v1alpha1",
              "kind":"TrafficRoute",
              "metadata":{
                "name":"empty",
                "creationTimestamp":null,
                "annotations": {
                  "k8s.kuma.io/synced": "true"
                }
              },
              "spec":{
                "sources":[
                  {
                    "match":{
                      "kuma.io/service":"web"
                    }
                  }
                ],
                "destinations":[
                  {
                    "match":{
                      "kuma.io/service":"backend"
                    }
                  }
                ],
                "conf":{
                 "split":[
                  {
                    "weight":100,
                    "destination":{
                      "kuma.io/service":"backend"
                    }
                  }
                ]
                }
              }
            }`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					UID:     "12345",
					Allowed: false,
					Result: &kube_meta.Status{
						Status:  "Failure",
						Message: "mesh: cannot be empty",
						Reason:  "Invalid",
						Code:    422,
						Details: &kube_meta.StatusDetails{
							Name: "empty",
							Kind: "TrafficRoute",
							Causes: []kube_meta.StatusCause{
								{
									Type:    "FieldValueInvalid",
									Message: "cannot be empty",
									Field:   "mesh",
								},
							},
						},
					},
				},
			},
			operation: admissionv1.Create,
		}),
		Entry("should fail validation on DELETE in Zone CP", testCase{
			mode:        core.Zone,
			objTemplate: &mesh_proto.TrafficRoute{},
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					UID:     "12345",
					Allowed: false,
					Result: &kube_meta.Status{
						Status:  "Failure",
						Message: "Operation not allowed. Kuma resources like TrafficRoute can be updated or deleted only from the GLOBAL control plane and not from a ZONE control plane.",
						Reason:  "Forbidden",
						Details: &kube_meta.StatusDetails{
							Causes: []kube_meta.StatusCause{
								{
									Type:    "FieldValueInvalid",
									Message: "cannot be empty",
									Field:   "metadata.annotations[kuma.io/synced]",
								},
							},
						},
						Code: 403,
					},
				},
			},
			operation: admissionv1.Delete,
		}),
		Entry("should fail validation on UPDATE in Zone CP", testCase{
			mode:        core.Zone,
			objTemplate: &mesh_proto.TrafficRoute{},
			obj: `
            {
              "apiVersion":"kuma.io/v1alpha1",
              "kind":"TrafficRoute",
              "mesh":"demo",
              "metadata":{
                "name":"empty",
                "creationTimestamp":null
              },
              "spec":{
                "sources":[
                  {
                    "match":{
                      "kuma.io/service":"web"
                    }
                  }
                ],
                "destinations":[
                  {
                    "match":{
                      "kuma.io/service":"backend"
                    }
                  }
                ],
                "conf":{
                 "split":[
                  {
                    "weight":100,
                    "destination":{
                      "kuma.io/service":"backend"
                    }
                  }
                ]
                }
              }
            }`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					UID:     "12345",
					Allowed: false,
					Result: &kube_meta.Status{
						Status:  "Failure",
						Message: "Operation not allowed. Kuma resources like TrafficRoute can be updated or deleted only from the GLOBAL control plane and not from a ZONE control plane.",
						Reason:  "Forbidden",
						Details: &kube_meta.StatusDetails{
							Causes: []kube_meta.StatusCause{
								{
									Type:    "FieldValueInvalid",
									Message: "cannot be empty",
									Field:   "metadata.annotations[kuma.io/synced]",
								},
							},
						},
						Code: 403,
					},
				},
			},
			operation: admissionv1.Update,
		}),
		Entry("should pass validation on DELETE in Global CP", testCase{
			mode:        core.Global,
			objTemplate: &mesh_proto.TrafficRoute{},
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					UID:     "12345",
					Allowed: true,
					Result: &kube_meta.Status{
						Code: 200,
					},
				},
			},
			operation: admissionv1.Delete,
		}),
	)
})
