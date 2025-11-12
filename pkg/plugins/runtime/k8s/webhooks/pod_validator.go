package webhooks

import (
	"context"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	kube_core "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
)

func NewPodValidatorWebhook(decoder admission.Decoder) *PodValidator {
	return &PodValidator{
		decoder: decoder,
	}
}

type PodValidator struct {
	decoder admission.Decoder
}

func (h *PodValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case v1.Create, v1.Update:
		return h.ValidatePod(ctx, req)
	}
	return admission.Allowed("")
}

func (h *PodValidator) ValidatePod(_ context.Context, req admission.Request) admission.Response {
	pod := &kube_core.Pod{}
	if err := h.decoder.Decode(req, pod); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Check if kuma.io/workload label is manually set
	if _, ok := pod.Labels[metadata.KumaWorkload]; ok {
		return admission.Denied("cannot manually set kuma.io/workload label on Pod; it is automatically managed by Kuma")
	}

	return admission.Allowed("")
}
