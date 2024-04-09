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

	config_core "github.com/kumahq/kuma/pkg/config/core"
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
<<<<<<< HEAD
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
=======
		inputObject            string
		expectedPatch          string
		expectedMessage        string
		skipMeshOwnerReference bool
		ownerId                kube_meta.Object
		cpMode                 string
	}
	DescribeTable("should add owner reference to resource owned by Mesh",
		func(given testCase) {
			if given.ownerId == nil {
				given.ownerId = defaultMesh
			}
			k8sGroupVersionKind := kube_meta.GroupVersionKind{}
			Expect(json.Unmarshal([]byte(given.inputObject), &k8sGroupVersionKind)).To(Succeed())
			wh := &webhooks.OwnerReferenceMutator{
				Client:                 k8sClient,
				CoreRegistry:           core_registry.Global(),
				K8sRegistry:            k8s_registry.Global(),
				Decoder:                decoder,
				Scheme:                 scheme,
				SkipMeshOwnerReference: given.skipMeshOwnerReference,
				CpMode:                 given.cpMode,
			}
			req := kube_admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: "12345",
					Object: kube_runtime.RawExtension{
						Raw: []byte(given.inputObject),
					},
					Kind: kube_meta.GroupVersionKind{
						Group:   k8sGroupVersionKind.Group,
						Version: k8sGroupVersionKind.Version,
						Kind:    k8sGroupVersionKind.Kind,
					},
				},
			}
			r := wh.Handle(context.Background(), req)
>>>>>>> 39495fb14 (feat(kuma-cp): do not set mesh owner reference on synced resources (#9882))
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
<<<<<<< HEAD
=======
		Entry("should not add patches to MeshService", testCase{
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "MeshService",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null
              }
            }`,
			expectedMessage: "ignored. MeshService has a reference for Service",
		}),
		Entry("should add owner reference to resource owned by Dataplane", testCase{
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "DataplaneInsight",
              "mesh": "default",
              "metadata": {
                "namespace": "default",
                "name": "dp-1",
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
                    "kind": "Dataplane",
                    "name": "dp-1",
                    "uid": "%s"
                  }
                ]
              }
            ]`,
			ownerId: dp1,
		}),
		Entry("should add owner reference to resource owned by Dataplane even with SkipMeshOwnerReference", testCase{
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "DataplaneInsight",
              "mesh": "default",
              "metadata": {
                "namespace": "default",
                "name": "dp-1",
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
                    "kind": "Dataplane",
                    "name": "dp-1",
                    "uid": "%s"
                  }
                ]
              }
            ]`,
			ownerId:                dp1,
			skipMeshOwnerReference: true,
		}),
		Entry("should ignore mesh owner reference", testCase{
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
			skipMeshOwnerReference: true,
			expectedMessage:        "ignored. Configuration setup to ignore Mesh owner reference.",
		}),
		Entry("should not add owner reference to synced resources to zone", testCase{
			cpMode: config_core.Zone,
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "TrafficRoute",
              "mesh": "default",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null,
                "labels": {
                  "kuma.io/origin": "global"
                }
              }
            }`,
			expectedMessage: "ignore. It's synced resource.",
		}),
		Entry("should not add owner reference to synced resources to global", testCase{
			cpMode: config_core.Global,
			inputObject: `
            {
              "apiVersion": "kuma.io/v1alpha1",
              "kind": "TrafficRoute",
              "mesh": "default",
              "metadata": {
                "namespace": "example",
                "name": "empty",
                "creationTimestamp": null,
                "labels": {
                  "kuma.io/origin": "zone"
                }
              }
            }`,
			expectedMessage: "ignore. It's synced resource.",
		}),
>>>>>>> 39495fb14 (feat(kuma-cp): do not set mesh owner reference on synced resources (#9882))
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
