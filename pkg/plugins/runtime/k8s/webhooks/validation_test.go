package webhooks_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/mode"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	k8s_resources "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
)

var _ = Describe("Validation", func() {

	var converter *k8s_resources.SimpleConverter

	BeforeEach(func() {
		converter = &k8s_resources.SimpleConverter{
			KubeFactory: &k8s_resources.SimpleKubeFactory{
				KubeTypes: k8s_registry.Global(),
			},
		}
	})

	type testCase struct {
		objTemplate core_model.ResourceSpec
		obj         string
		mode        mode.CpMode
		resp        kube_admission.Response
	}
	DescribeTable("Validation",
		func(given testCase) {
			// given
			webhook := &admission.Webhook{
				Handler: webhooks.NewValidatingWebhook(converter, core_registry.Global(), k8s_registry.Global(), given.mode),
			}
			Expect(webhook.InjectScheme(scheme)).To(Succeed())

			obj, err := k8s_registry.Global().NewObject(given.objTemplate)
			Expect(err).ToNot(HaveOccurred())
			req := kube_admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					UID: kube_types.UID("12345"),
					Object: kube_runtime.RawExtension{
						Raw: []byte(given.obj),
					},
					Kind: kube_meta.GroupVersionKind{
						Group:   obj.GetObjectKind().GroupVersionKind().Group,
						Version: obj.GetObjectKind().GroupVersionKind().Version,
						Kind:    obj.GetObjectKind().GroupVersionKind().Kind,
					},
				},
			}

			// when
			resp := webhook.Handle(context.Background(), req)

			// then
			Expect(resp).To(Equal(given.resp))
		},
		Entry("should pass validation", testCase{
			mode:        mode.Standalone,
			objTemplate: &mesh_proto.TrafficRoute{},
			obj: `
            {
              "apiVersion":"kuma.io/v1alpha1",
              "kind":"TrafficRoute",
              "mesh":"demo",
              "metadata":{
                "namespace":"example",
                "name":"empty",
                "creationTimestamp":null
              },
              "spec":{
                "sources":[
                  {
                    "match":{
                      "service":"web"
                    }
                  }
                ],
                "destinations":[
                  {
                    "match":{
                      "service":"backend"
                    }
                  }
                ],
                "conf":[
                  {
                    "weight":100,
                    "destination":{
                      "service":"backend"
                    }
                  }
                ]
              }
            }`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1beta1.AdmissionResponse{
					UID:     "12345",
					Allowed: true,
					Result: &kube_meta.Status{
						Code: 200,
					},
				},
			},
		}),
		Entry("should pass default mesh on remote", testCase{
			mode:        mode.Remote,
			objTemplate: &mesh_proto.Mesh{},
			obj: `
            {
              "apiVersion":"kuma.io/v1alpha1",
              "kind":"Mesh",
              "mesh":"default",
              "metadata":{
                "name":"default",
                "creationTimestamp":null
              },
              "spec":{}
            }`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1beta1.AdmissionResponse{
					UID:     "12345",
					Allowed: true,
					Result: &kube_meta.Status{
						Code: 200,
					},
				},
			},
		}),
		Entry("should pass validation for synced policy from Global to Remote", testCase{
			mode:        mode.Remote,
			objTemplate: &mesh_proto.TrafficRoute{},
			obj: `
            {
              "apiVersion":"kuma.io/v1alpha1",
              "kind":"TrafficRoute",
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
                "sources":[
                  {
                    "match":{
                      "service":"web"
                    }
                  }
                ],
                "destinations":[
                  {
                    "match":{
                      "service":"backend"
                    }
                  }
                ],
                "conf":[
                  {
                    "weight":100,
                    "destination":{
                      "service":"backend"
                    }
                  }
                ]
              }
            }`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1beta1.AdmissionResponse{
					UID:     "12345",
					Allowed: true,
					Result: &kube_meta.Status{
						Code: 200,
					},
				},
			},
		}),
		Entry("should pass validation for synced policy from Remote to Global", testCase{
			mode:        mode.Remote,
			objTemplate: &mesh_proto.Dataplane{},
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
                        "service": "backend"
                      }
                    }
                  ]
                }
              }
            }`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1beta1.AdmissionResponse{
					UID:     "12345",
					Allowed: true,
					Result: &kube_meta.Status{
						Code: 200,
					},
				},
			},
		}),
		Entry("should pass validation for not synced Dataplane in Remote", testCase{
			mode:        mode.Remote,
			objTemplate: &mesh_proto.Dataplane{},
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
                        "service": "backend"
                      }
                    }
                  ]
                }
              }
            }`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1beta1.AdmissionResponse{
					UID:     "12345",
					Allowed: true,
					Result: &kube_meta.Status{
						Code: 200,
					},
				},
			},
		}),
		Entry("should fail validation due to invalid spec", testCase{
			mode:        mode.Global,
			objTemplate: &mesh_proto.TrafficRoute{},
			obj: `
			{
			  "apiVersion": "kuma.io/v1alpha1",
			  "kind": "TrafficRoute",
			  "mesh": "demo",
			  "metadata": {
				"namespace": "example",
				"name": "empty",
				"creationTimestamp": null
			  },
			  "spec": {
			  }
			}
			`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1beta1.AdmissionResponse{
					UID:     "12345",
					Allowed: false,
					Result: &kube_meta.Status{
						Status:  "Failure",
						Message: "spec.sources: must have at least one element; spec.destinations: must have at least one element; spec.conf: must have at least one element",
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
									Message: "must have at least one element",
									Field:   "spec.conf",
								},
							},
						},
						Code: 422,
					},
				},
			},
		}),
		Entry("should fail validation due to applying policy manually on Remote CP", testCase{
			mode:        mode.Remote,
			objTemplate: &mesh_proto.TrafficRoute{},
			obj: `
			{
			  "apiVersion": "kuma.io/v1alpha1",
			  "kind": "TrafficRoute",
			  "mesh": "demo",
			  "metadata": {
				"namespace": "example",
				"name": "empty",
				"creationTimestamp": null
			  }
			}
			`,
			resp: kube_admission.Response{
				AdmissionResponse: admissionv1beta1.AdmissionResponse{
					UID:     "12345",
					Allowed: false,
					Result: &kube_meta.Status{
						Status:  "Failure",
						Message: "You are trying to apply a TrafficRoute on remote CP. In multicluster setup, it should be only applied on global CP and synced to remote CP.",
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
		}),
		Entry("should fail validation due to applying Dataplane manually on Global CP", testCase{
			mode:        mode.Global,
			objTemplate: &mesh_proto.Dataplane{},
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
				AdmissionResponse: admissionv1beta1.AdmissionResponse{
					UID:     "12345",
					Allowed: false,
					Result: &kube_meta.Status{
						Status:  "Failure",
						Message: "You are trying to apply a Dataplane on global CP. In multicluster setup, it should be only applied on remote CP and synced to global CP.",
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
		}),
	)
})
