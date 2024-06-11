package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
)

type Defaulter interface {
	core_model.Resource
	Default() error
}

func DefaultingWebhookFor(scheme *runtime.Scheme, converter k8s_common.Converter, checker ResourceAdmissionChecker) *admission.Webhook {
	return &admission.Webhook{
		Handler: &defaultingHandler{
			converter:                converter,
			decoder:                  admission.NewDecoder(scheme),
			ResourceAdmissionChecker: checker,
		},
	}
}

type defaultingHandler struct {
	ResourceAdmissionChecker

	converter k8s_common.Converter
	decoder   *admission.Decoder
}

func (h *defaultingHandler) Handle(_ context.Context, req admission.Request) admission.Response {
	resource, err := registry.Global().NewObject(core_model.ResourceType(req.Kind.Kind))
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

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

	if resp := h.IsOperationAllowed(req.UserInfo, resource, req.Namespace); !resp.Allowed {
		return resp
	}
	labels, annotations := k8s.SplitLabelsAndAnnotations(
<<<<<<< HEAD
		core_model.ComputeLabels(resource, h.Mode, true, h.SystemNamespace, h.ZoneName),
=======
		core_model.ComputeLabels(resource, h.Mode, true, h.SystemNamespace),
>>>>>>> da824ce57 (fix(kuma-cp): mistakenly setting 'kuma.io/display-name' as label (#10430))
		obj.GetAnnotations(),
	)
	obj.SetLabels(labels)
	obj.SetAnnotations(annotations)

	marshaled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}
