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
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_resources "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	sample_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/test/api/sample/v1alpha1"
	. "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
	sample_proto "github.com/kumahq/kuma/pkg/test/apis/sample/v1alpha1"
	sample_core "github.com/kumahq/kuma/pkg/test/resources/apis/sample"
)

var _ = Describe("Defaulter", func() {

	var factory func() core_model.Resource

	BeforeEach(func() {
		factory = func() core_model.Resource {
			return sample_core.NewTrafficRouteResource()
		}
	})

	var converter k8s_common.Converter

	BeforeEach(func() {
		kubeTypes := k8s_registry.NewTypeRegistry()
		err := kubeTypes.RegisterObjectType(&sample_proto.TrafficRoute{}, &sample_k8s.SampleTrafficRoute{
			TypeMeta: kube_meta.TypeMeta{
				APIVersion: sample_k8s.GroupVersion.String(),
				Kind:       "SampleTrafficRoute",
			},
		})
		Expect(err).ToNot(HaveOccurred())

		converter = &k8s_resources.SimpleConverter{
			KubeFactory: &k8s_resources.SimpleKubeFactory{
				KubeTypes: kubeTypes,
			},
		}
	})

	var scheme *kube_runtime.Scheme

	BeforeEach(func() {
		scheme = kube_runtime.NewScheme()
		Expect(sample_k8s.AddToScheme(scheme)).To(Succeed())
	})

	var handler *kube_admission.Webhook

	BeforeEach(func() {
		handler = DefaultingWebhookFor(factory, converter)
		Expect(handler.InjectScheme(scheme)).To(Succeed())
	})

	type testCase struct {
		inputObject   string
		expectedPatch string
	}

	DescribeTable("should apply defaults on a target object",
		func(given testCase) {
			// given
			req := kube_admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					UID: kube_types.UID("12345"),
					Object: kube_runtime.RawExtension{
						Raw: []byte(given.inputObject),
					},
				},
			}

			// when
			resp := handler.Handle(context.Background(), req)

			// then
			Expect(resp.UID).To(Equal(kube_types.UID("12345")))
			Expect(resp.Result.Message).To(Equal(""))
			Expect(resp.Allowed).To(Equal(true))
			Expect(*resp.PatchType).To(Equal(admissionv1beta1.PatchTypeJSONPatch))
			if given.expectedPatch != "" {
				Expect(string(resp.Patch)).To(MatchJSON(given.expectedPatch))
			} else {
				Expect(string(resp.Patch)).To(BeZero())
			}
		},
		Entry("should apply defaults to empty spec fields", testCase{
			inputObject: `
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
			expectedPatch: `
            [
              {
                "op": "add",
                "path": "/spec/path",
                "value": "/default"
              }
            ]
`,
		}),
		Entry("should not override non-empty spec fields", testCase{
			inputObject: `
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
                "path": "anything"
              }
            }
`,
			expectedPatch: ``,
		}),
		Entry("should not override spec fields already set to default value", testCase{
			inputObject: `
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
                "path": "/default"
              }
            }
`,
			expectedPatch: ``,
		}),
	)
})
