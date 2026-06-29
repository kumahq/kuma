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
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
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

	decision := h.IsOperationAllowed(req.UserInfo, resource, req.Namespace, req.Operation)
	if !decision.Response.Allowed {
		return decision.Response
	}

	coreLabels := resource.GetMeta().GetLabels()
	k8sName := resource.GetMeta().GetName()
	if name, ok := resource.GetMeta().GetNameExtensions()[core_model.K8sNameComponent]; ok && name != "" {
		k8sName = name
	}
	mesh := resource.GetMeta().GetMesh()
	if labelMesh, ok := coreLabels[metadata.KumaMeshLabel]; ok && labelMesh != "" {
		mesh = labelMesh
	}
	previousLabels, err := h.previousLabels(req)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	computed, err := resource_labels.Compute(
		resource.Descriptor(),
		resource.GetSpec(),
		coreLabels,
		mesh,
		k8sName,
		resource_labels.WithNamespace(resource_labels.GetNamespace(resource.GetMeta(), h.SystemNamespace)),
		resource_labels.WithMode(h.Mode),
		resource_labels.WithK8s(true),
		resource_labels.WithZone(h.ZoneName),
		resource_labels.WithPrivileged(decision.Privileged),
		resource_labels.WithPreviousLabels(previousLabels),
	)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	coreLabels = computed
	labels, annotations := k8s.SplitLabelsAndAnnotations(coreLabels, obj.GetAnnotations())

	obj.SetLabels(labels)
	obj.SetAnnotations(annotations)

	marshaled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled).WithWarnings(decision.Warnings...)
}

func (h *defaultingHandler) previousLabels(req admission.Request) (map[string]string, error) {
	if req.Operation != admissionv1.Update || len(req.OldObject.Raw) == 0 {
		return nil, nil
	}
	oldResource, err := registry.Global().NewObject(core_model.ResourceType(req.Kind.Kind))
	if err != nil {
		return nil, err
	}
	oldObj, err := h.converter.ToKubernetesObject(oldResource)
	if err != nil {
		return nil, err
	}
	if err := h.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
		return nil, err
	}
	if err := h.converter.ToCoreResource(oldObj, oldResource); err != nil {
		return nil, err
	}
	return oldResource.GetMeta().GetLabels(), nil
}
