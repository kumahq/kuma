package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	k8s_common "github.com/kumahq/kuma/v3/pkg/plugins/common/k8s"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s"
	k8s_model "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/model"
)

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
	decoder   admission.Decoder
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

	if defaulter, ok := resource.(core_model.Defaulter); ok {
		if err := defaulter.Default(); err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	obj, err = h.converter.ToKubernetesObject(resource)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	var previous k8s_model.KubernetesObject
	if req.Operation == admissionv1.Update {
		previous, err = h.decodeObject(req.Kind.Kind, req.OldObject)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
	}
	if resp := h.IsOperationAllowed(req.UserInfo, resource, obj, previous, req.Namespace); !resp.Allowed {
		return resp
	}

	displayName := resource.GetMeta().GetName()
	if name, ok := resource.GetMeta().GetNameExtensions()[core_model.K8sNameComponent]; ok && name != "" {
		displayName = name
	}
	// Compute only fails on a policy the user got wrong (mixed producer and consumer
	// items), so it is forbidden rather than an internal error.
	computed, err := resource_labels.Compute(resource_labels.Write{
		Descriptor:    resource.Descriptor(),
		Spec:          resource.GetSpec(),
		Namespace:     resource_labels.GetNamespace(resource.GetMeta(), h.SystemNamespace),
		Mesh:          resource.GetMeta().GetMesh(),
		DisplayName:   displayName,
		Labels:        resource.GetMeta().GetLabels(),
		TrustedWriter: h.isPrivilegedUser(h.AllowedUsers, req.UserInfo),
	}, h.ControlPlane)
	if err != nil {
		return *forbiddenResponse(err.Error())
	}
	labels, annotations := k8s.SplitLabelsAndAnnotations(computed, obj.GetAnnotations())

	obj.SetLabels(labels)
	obj.SetAnnotations(annotations)

	marshaled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}

func (h *defaultingHandler) decodeObject(kind string, raw runtime.RawExtension) (k8s_model.KubernetesObject, error) {
	resource, err := registry.Global().NewObject(core_model.ResourceType(kind))
	if err != nil {
		return nil, err
	}
	obj, err := h.converter.ToKubernetesObject(resource)
	if err != nil {
		return nil, err
	}
	if err := h.decoder.DecodeRaw(raw, obj); err != nil {
		return nil, err
	}
	return obj, nil
}
