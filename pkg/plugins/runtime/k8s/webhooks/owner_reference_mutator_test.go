package webhooks_test

import (
	"context"
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
)

var _ = Describe("OwnerReferenceMutator", func() {

	createWebhook := func() webhook.AdmissionHandler {
		return &webhooks.OwnerReferenceMutator{
			Client:       k8sClient,
			CoreRegistry: core_registry.Global(),
			K8sRegistry:  k8s_registry.Global(),
			Decoder:      decoder,
			Scheme:       scheme,
		}
	}

	createRequest := func(obj model.KubernetesObject, raw []byte) kube_admission.Request {
		return kube_admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				UID: "12345",
				Object: kube_runtime.RawExtension{
					Raw: raw,
				},
				Kind: kube_meta.GroupVersionKind{
					Group:   obj.GetObjectKind().GroupVersionKind().Group,
					Version: obj.GetObjectKind().GroupVersionKind().Version,
					Kind:    obj.GetObjectKind().GroupVersionKind().Kind,
				},
			},
		}
	}

	type testCase struct {
		inputObject     string
		expectedPatch   string
		expectedMessage string
	}
	DescribeTable("should add owner reference to resource owned by Mesh",
		func(given testCase) {
			tr := &mesh_k8s.TrafficRoute{}
			err := json.Unmarshal([]byte(given.inputObject), tr)
			Expect(err).ToNot(HaveOccurred())

			wh := createWebhook()
			r := wh.Handle(context.Background(), createRequest(tr, []byte(given.inputObject)))
			if given.expectedMessage != "" {
				Expect(r.Result.Message).To(Equal(given.expectedMessage))
			} else {
				Expect(r.Result).To(BeNil())
			}
			if given.expectedPatch != "" {
				patch, err := json.Marshal(r.Patches)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(patch)).To(MatchJSON(fmt.Sprintf(given.expectedPatch, defaultMesh.GetUID())))
			} else {
				Expect(r.Patches).To(BeNil())
			}
		},
		Entry("should add response patch", testCase{
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "TrafficRoute",
              "mesh": "default",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null
              }
            }`,
			expectedPatch: `
            [
              {
                "op": "add",
                "path": "/metadata/ownerReferences",
                "value": [
                  {
                    "apiVersion": "kuma.io/v1alpha1",
                    "kind": "Mesh",
                    "name": "default",
                    "uid": "%s"
                  }
                ]
              }
            ]`,
		}),
		Entry("should return error message if mesh doesn't exist", testCase{
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "TrafficRoute",
              "mesh": "not-existing-mesh",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null
              }
            }`,
			expectedMessage: `meshes.kuma.io "not-existing-mesh" not found`,
		}),
		Entry("should return error message if mesh is not present", testCase{
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "TrafficRoute",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null
              }
            }`,
			expectedMessage: `mesh: cannot be empty`,
		}),
	)

	It("should add owner reference to resource owned by Dataplane", func() {
		inputObject := `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "DataplaneInsight",
              "mesh": "default",
              "metadata": {
                "namespace": "default",
                "name": "dp-1",
                "creationTimestamp": null
              }
            }`
		dpInsight := &mesh_k8s.DataplaneInsight{}
		err := json.Unmarshal([]byte(inputObject), dpInsight)
		Expect(err).ToNot(HaveOccurred())

		dp := &mesh_k8s.Dataplane{
			ObjectMeta: kube_meta.ObjectMeta{
				Name:      "dp-1",
				Namespace: "default",
			},
			Mesh: "default",
		}
		err = k8sClient.Create(context.Background(), dp)
		Expect(err).ToNot(HaveOccurred())

		wh := createWebhook()
		r := wh.Handle(context.Background(), createRequest(dpInsight, []byte(inputObject)))

		expectedPatch := fmt.Sprintf(`
            [
              {
                "op": "add",
                "path": "/metadata/ownerReferences",
                "value": [
                  {
                    "apiVersion": "kuma.io/v1alpha1",
                    "kind": "Dataplane",
                    "name": "dp-1",
                    "uid": "%s"
                  }
                ]
              }
            ]`, dp.GetUID())
		patch, err := json.Marshal(r.Patches)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(patch)).To(MatchJSON(expectedPatch))
	})

})
