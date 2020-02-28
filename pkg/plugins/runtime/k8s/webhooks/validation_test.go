package webhooks_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	core_registry "github.com/Kong/kuma/pkg/core/resources/registry"
	k8s_resources "github.com/Kong/kuma/pkg/plugins/resources/k8s"
	k8s_registry "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	sample_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/test/api/sample/v1alpha1"
	"github.com/Kong/kuma/pkg/plugins/runtime/k8s/webhooks"
	sample_proto "github.com/Kong/kuma/pkg/test/apis/sample/v1alpha1"
	sample_core "github.com/Kong/kuma/pkg/test/resources/apis/sample"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ = Describe("Validation", func() {

	var handler *admission.Webhook
	var kubeTypes k8s_registry.TypeRegistry

	BeforeEach(func() {
		kubeTypes = k8s_registry.NewTypeRegistry()
		err := kubeTypes.RegisterObjectType(&sample_proto.TrafficRoute{}, &sample_k8s.SampleTrafficRoute{
			TypeMeta: kube_meta.TypeMeta{
				APIVersion: sample_k8s.GroupVersion.String(),
				Kind:       "SampleTrafficRoute",
			},
		})
		Expect(err).ToNot(HaveOccurred())

		converter := &k8s_resources.SimpleConverter{
			KubeFactory: &k8s_resources.SimpleKubeFactory{
				KubeTypes: kubeTypes,
			},
		}

		types := core_registry.NewTypeRegistry()
		err = types.RegisterType(&sample_core.TrafficRouteResource{})
		Expect(err).ToNot(HaveOccurred())

		webhook := &admission.Webhook{
			Handler: webhooks.NewValidatingWebhook(converter, types, kubeTypes),
		}

		scheme := kube_runtime.NewScheme()
		Expect(sample_k8s.AddToScheme(scheme)).To(Succeed())
		Expect(webhook.InjectScheme(scheme)).To(Succeed())
		handler = webhook
	})

	type testCase struct {
		obj  string
		resp kube_admission.Response
	}
	DescribeTable("Validation",
		func(given testCase) {
			// given
			obj, err := kubeTypes.NewObject(&sample_proto.TrafficRoute{})
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
			resp := handler.Handle(context.Background(), req)

			// then
			Expect(resp).To(Equal(given.resp))
		},
		Entry("should pass validation", testCase{
			obj: `
			{
			  "apiVersion": "sample.test.kuma.io/v1alpha1",
			  "kind": "SampleTrafficRoute",
			  "mesh": "demo",
			  "metadata": {
				"namespace": "example",
				"name": "empty",
				"creationTimestamp": null
			  },
			  "spec": {
				"path": "/random"
			  }
			}
			`,
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
		Entry("should fail validation", testCase{
			obj: `
			{
			  "apiVersion": "sample.test.kuma.io/v1alpha1",
			  "kind": "SampleTrafficRoute",
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
						Message: "spec.path: cannot be empty",
						Reason:  "Invalid",
						Details: &kube_meta.StatusDetails{
							Name: "empty",
							Kind: "SampleTrafficRoute",
							Causes: []kube_meta.StatusCause{
								{
									Type:    "FieldValueInvalid",
									Message: "cannot be empty",
									Field:   "spec.path",
								},
							},
						},
						Code: 422,
					},
				},
			},
		}),
	)
})
