package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	k8s_resources "github.com/Kong/kuma/pkg/plugins/resources/k8s"
)

type Defaulter interface {
	core_model.Resource
	Default() error
}

func DefaultingWebhookFor(factory func() core_model.Resource, converter k8s_resources.Converter) *admission.Webhook {
	return &admission.Webhook{
		Handler: &defaultingHandler{
			factory:   factory,
			converter: converter,
		},
	}
}

type defaultingHandler struct {
	factory   func() core_model.Resource
	converter k8s_resources.Converter
	decoder   *admission.Decoder
}

var _ admission.DecoderInjector = &defaultingHandler{}

func (h *defaultingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *defaultingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	resource := h.factory()

	obj, err := h.converter.ToKubernetesObject(resource)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	err = h.decoder.Decode(req, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := h.converter.ToCoreResource(obj, resource); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if defaulter, ok := resource.(Defaulter); ok {
		if err := defaulter.Default(); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	obj, err = h.converter.ToKubernetesObject(resource)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	marshalled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshalled)
}
