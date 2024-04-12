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
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
)

var _ = Describe("OwnerReferenceMutator", func() {
	type testCase struct {
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
			if given.expectedMessage != "" {
				Expect(r.Result.Message).To(Equal(given.expectedMessage))
			} else {
				Expect(r.Result).To(BeNil())
			}
			if given.expectedPatch != "" {
				patch, err := json.Marshal(r.Patches)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(patch)).To(MatchJSON(fmt.Sprintf(given.expectedPatch, given.ownerId.GetUID())))
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
	)
})
