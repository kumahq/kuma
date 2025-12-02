package webhooks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/admission/v1"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_k8s "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/util"
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
	ns := &kube_core.Namespace{}
	if err := h.client.Get(ctx, kube_client.ObjectKey{Name: pod.Namespace}, ns); err != nil {
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to get namespace: %w", err))
	}

	podMesh := util.MeshOfByLabelOrAnnotation(logr.Discard(), pod, ns)

	// Create a selector to find dataplanes NOT in our mesh
	// This is much more efficient than listing all dataplanes and filtering
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(
		metadata.KumaMeshLabel,
		selection.NotEquals,
		[]string{podMesh},
	)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to create label selector: %w", err))
	}
	selector = selector.Add(*requirement)

	// List only dataplanes in different meshes (with limit=1 since we only need one to reject)
	dataplanes := &mesh_k8s.DataplaneList{}
	if err := h.client.List(ctx, dataplanes,
		kube_client.InNamespace(pod.Namespace),
		kube_client.MatchingLabelsSelector{Selector: selector},
		kube_client.Limit(1),
	); err != nil {
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to list dataplanes: %w", err))
	}

	// If we found any dataplane in a different mesh, deny
	if len(dataplanes.Items) > 0 {
		return admission.Denied(fmt.Sprintf(
			"pod is in mesh %q but namespace %q already contains dataplanes in mesh %q; only one mesh per namespace is allowed when runtime.kubernetes.disallowMultipleMeshesPerNamespace is enabled",
			podMesh, pod.Namespace, dataplanes.Items[0].Mesh,
		))
	}

	return admission.Allowed("")
}
