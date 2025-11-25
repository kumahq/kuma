package webhooks

import (
	"context"
	"fmt"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_k8s "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
)

func NewPodValidatorWebhook(decoder admission.Decoder, client kube_client.Client, disallowMultipleMeshesPerNamespace bool) *PodValidator {
	return &PodValidator{
		decoder:                            decoder,
		client:                             client,
		disallowMultipleMeshesPerNamespace: disallowMultipleMeshesPerNamespace,
	}
}

type PodValidator struct {
	decoder                            admission.Decoder
	client                             kube_client.Client
	disallowMultipleMeshesPerNamespace bool
}

func (h *PodValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case v1.Create, v1.Update:
		return h.ValidatePod(ctx, req)
	}
	return admission.Allowed("")
}

func (h *PodValidator) ValidatePod(ctx context.Context, req admission.Request) admission.Response {
	pod := &kube_core.Pod{}
	if err := h.decoder.Decode(req, pod); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Check if kuma.io/workload label is manually set
	if _, ok := pod.Labels[metadata.KumaWorkload]; ok {
		return admission.Denied("cannot manually set kuma.io/workload label on Pod; it is automatically managed by Kuma")
	}

	// Check for multiple meshes per namespace if the feature is enabled
	if h.disallowMultipleMeshesPerNamespace {
		if resp := h.validateMultipleMeshesPerNamespace(ctx, pod); !resp.Allowed {
			return resp
		}
	}

	return admission.Allowed("")
}

func (h *PodValidator) validateMultipleMeshesPerNamespace(ctx context.Context, pod *kube_core.Pod) admission.Response {
	// Get the mesh for this pod
	podMesh := pod.Annotations[metadata.KumaMeshLabel]
	if podMesh == "" {
		podMesh = "default"
	}

	// List all dataplanes in the namespace to check for different meshes
	dataplanes := &mesh_k8s.DataplaneList{}
	if err := h.client.List(ctx, dataplanes, kube_client.InNamespace(pod.Namespace)); err != nil {
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to list dataplanes: %w", err))
	}

	// Check if there are dataplanes in a different mesh
	for _, dp := range dataplanes.Items {
		if dp.Mesh != podMesh {
			return admission.Denied(fmt.Sprintf(
				"pod is in mesh %q but namespace %q already contains dataplanes in mesh %q; only one mesh per namespace is allowed when runtime.kubernetes.disallowMultipleMeshesPerNamespace is enabled",
				podMesh, pod.Namespace, dp.Mesh,
			))
		}
	}

	return admission.Allowed("")
}
