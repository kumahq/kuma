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

	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_resources "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	sample_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/test/api/sample/v1alpha1"
	. "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
	sample_proto "github.com/kumahq/kuma/pkg/test/apis/sample/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/apis/sample"
)

var _ = Describe("Defaulter", func() {

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
		handler = DefaultingWebhookFor(converter)
		Expect(handler.InjectScheme(scheme)).To(Succeed())
	})

	type testCase struct {
		inputObject string
		expected    string
	}

	DescribeTable("should apply defaults on a target object",
		func(given testCase) {
			// given
			req := kube_admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: kube_types.UID("12345"),
					Object: kube_runtime.RawExtension{
						Raw: []byte(given.inputObject),
					},
					Kind: kube_meta.GroupVersionKind{
						Kind: string(sample.TrafficRouteType),
					},
				},
			}

			// when
			resp := handler.Handle(context.Background(), req)

			// then
			Expect(resp.UID).To(Equal(kube_types.UID("12345")))
			Expect(resp.Result.Message).To(Equal(""))
			Expect(resp.Allowed).To(Equal(true))

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
			expected: `
            {
              "apiVersion": "sample.test.kuma.io/v1alpha1",
              "kind": "SampleTrafficRoute",
              "mesh": "demo",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null,
                "labels": {
                  "kuma.io/mesh": "default"
                }
              },
              "spec": {
                "path": "/default"
              }
			}
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
			expected: `
            {
              "apiVersion": "sample.test.kuma.io/v1alpha1",
              "kind": "SampleTrafficRoute",
              "mesh": "demo",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null,
                "labels": {
                  "kuma.io/mesh": "default"
                }
              },
              "spec": {
                "path": "anything"
              }
            }
`,
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
			expected: `
            {
              "apiVersion": "sample.test.kuma.io/v1alpha1",
              "kind": "SampleTrafficRoute",
              "mesh": "demo",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null,
                "labels": {
                  "kuma.io/mesh": "default"
                }
              },
              "spec": {
                "path": "/default"
              }
            }
`,
		}),
		Entry("should not override mesh label if it's already set", testCase{
			inputObject: `
            {
              "apiVersion": "sample.test.kuma.io/v1alpha1",
              "kind": "SampleTrafficRoute",
              "mesh": "demo",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null,
                "labels": {
                  "kuma.io/mesh": "my-mesh-1"
                }
              },
              "spec": {
                "path": "/default"
              }
            }
`,
			expected: `
            {
              "apiVersion": "sample.test.kuma.io/v1alpha1",
              "kind": "SampleTrafficRoute",
              "mesh": "demo",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null,
                "labels": {
                  "kuma.io/mesh": "my-mesh-1"
                }
              },
              "spec": {
                "path": "/default"
              }
            }
`,
		}),
	)
})
