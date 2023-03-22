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

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_resources "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	. "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
)

var _ = Describe("Defaulter", func() {
	var converter k8s_common.Converter

	BeforeEach(func() {
		kubeTypes := k8s_registry.NewTypeRegistry()
		err := kubeTypes.RegisterObjectType(&mesh_proto.Mesh{}, &mesh_k8s.Mesh{
			TypeMeta: kube_meta.TypeMeta{
				APIVersion: mesh_k8s.GroupVersion.String(),
				Kind:       "Mesh",
			},
		})
		Expect(err).ToNot(HaveOccurred())
		err = kubeTypes.RegisterObjectType(&mesh_proto.TrafficRoute{}, &mesh_k8s.TrafficRoute{
			TypeMeta: kube_meta.TypeMeta{
				APIVersion: mesh_k8s.GroupVersion.String(),
				Kind:       "TrafficRoute",
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
		Expect(mesh_k8s.AddToScheme(scheme)).To(Succeed())
	})

	var handler *kube_admission.Webhook

	BeforeEach(func() {
		handler = DefaultingWebhookFor(converter)
		Expect(handler.InjectScheme(scheme)).To(Succeed())
	})

	type testCase struct {
		inputObject string
		expected    string
		kind        string
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
			kind: string(mesh.MeshType),
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
				"creationTimestamp": null
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
						}
					  }
					}
				  ]
				}
              }
            }
`,
		}),
		Entry("should not override non-empty spec fields", testCase{
			kind: string(mesh.MeshType),
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
				"creationTimestamp": null
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
						}
					  }
					}
				  ]
				}
              }
            }
`,
		}),
		Entry("should not override mesh label if it's already set", testCase{
			kind: string(mesh.TrafficRouteType),
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
                  "kuma.io/mesh": "my-mesh-1"
                }
              },
              "spec": {}
            }
`,
		}),
	)
})
