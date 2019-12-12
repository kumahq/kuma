package webhooks_test

import (
	"context"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_registry "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	sample_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/test/api/sample/v1alpha1"
	"github.com/Kong/kuma/pkg/plugins/runtime/k8s/webhooks"
	sample_proto "github.com/Kong/kuma/pkg/test/apis/sample/v1alpha1"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type denyingValidator struct {
}

var _ webhooks.AdmissionValidator = &denyingValidator{}

func (d *denyingValidator) InjectDecoder(*admission.Decoder) error {
	return nil
}

func (d *denyingValidator) Handle(context.Context, admission.Request) admission.Response {
	return admission.Denied("")
}

func (d *denyingValidator) Supports(req admission.Request) bool {
	gvk := sample_k8s.GroupVersion.WithKind("SampleTrafficRoute")
	return req.Kind.Group == gvk.Group && req.Kind.Kind == gvk.Kind && req.Kind.Version == gvk.Version
}

var _ = Describe("Composite Validator", func() {

	var handler admission.Handler
	var kubeTypes k8s_registry.TypeRegistry

	BeforeEach(func() {
		composite := webhooks.CompositeValidator{}
		composite.AddValidator(&denyingValidator{})

		handler = composite.WebHook()

		kubeTypes = k8s_registry.NewTypeRegistry()
		err := kubeTypes.RegisterObjectType(&sample_proto.TrafficRoute{}, &sample_k8s.SampleTrafficRoute{
			TypeMeta: kube_meta.TypeMeta{
				APIVersion: sample_k8s.GroupVersion.String(),
				Kind:       "SampleTrafficRoute",
			},
		})
		Expect(err).ToNot(HaveOccurred())

		err = kubeTypes.RegisterObjectType(&mesh_proto.Mesh{}, &mesh_k8s.Mesh{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should validate supported resource", func() {
		// given
		yaml := `
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
			`
		obj, err := kubeTypes.NewObject(&sample_proto.TrafficRoute{})
		Expect(err).ToNot(HaveOccurred())

		req := kube_admission.Request{
			AdmissionRequest: admissionv1beta1.AdmissionRequest{
				UID: kube_types.UID("12345"),
				Object: kube_runtime.RawExtension{
					Raw: []byte(yaml),
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
		Expect(resp.Allowed).To(BeFalse())
	})

	It("should skip validation for not supported resource", func() {
		// given
		yaml := `
			{
			  "apiVersion": "kuma.io/v1alpha1",
			  "kind": "Mesh",
			  "metadata": {
				"name": "empty",
				"creationTimestamp": null
			  },
			  "spec": {
			  }
			}
			`
		obj, err := kubeTypes.NewObject(&mesh_proto.Mesh{})
		Expect(err).ToNot(HaveOccurred())

		req := kube_admission.Request{
			AdmissionRequest: admissionv1beta1.AdmissionRequest{
				UID: kube_types.UID("12345"),
				Object: kube_runtime.RawExtension{
					Raw: []byte(yaml),
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
		Expect(resp.Allowed).To(BeTrue())
	})
})
