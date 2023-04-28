package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

type Defaulter interface {
	core_model.Resource
	Default() error
}

func DefaultingWebhookFor(reg registry.TypeRegistry, converter k8s_common.Converter) *admission.Webhook {
	return &admission.Webhook{
		Handler: &defaultingHandler{
			converter: converter,
			registry:  reg,
		},
	}
}

type defaultingHandler struct {
	converter k8s_common.Converter
	decoder   *admission.Decoder
	registry  registry.TypeRegistry
}

var _ admission.DecoderInjector = &defaultingHandler{}

func (h *defaultingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *defaultingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	desc, err := h.registry.DescriptorFor(core_model.ResourceType(req.Kind.Kind))
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	resource := desc.NewObject()

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

	if desc.Scope == core_model.ScopeMesh {
		labels := obj.GetLabels()
		if _, ok := labels[metadata.KumaMeshLabel]; !ok {
			if len(labels) == 0 {
				labels = map[string]string{}
			}
			labels[metadata.KumaMeshLabel] = core_model.DefaultMesh
			obj.SetLabels(labels)
		}
	}

	marshaled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}
