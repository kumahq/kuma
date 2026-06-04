package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	resource_labels "github.com/kumahq/kuma/v2/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
	k8s_common "github.com/kumahq/kuma/v2/pkg/plugins/common/k8s"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
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

	// Compute is the single mutator for labels on K8s. It uses the Privileged
	// signal to distinguish a KDS sync / GC / migrator write (non-locally
	// originated — labels are authoritative from the source CP) from a local
	// K8s controller write (locally originated — needs the registry walk to
	// fill in kuma.io/origin, kuma.io/namespace, kuma.io/display-name, ...
	// while preserving system labels like kuma.io/managed-by).
	coreLabels := resource.GetMeta().GetLabels()
	// The K8s adapter encodes the metadata.name into NameExtensions; fall
	// back to stripping the namespace suffix from the kuma name if the
	// extension is missing.
	k8sName := resource.GetMeta().GetName()
	if name, ok := resource.GetMeta().GetNameExtensions()[core_model.K8sNameComponent]; ok && name != "" {
		k8sName = name
	}
	// On K8s the kuma.io/mesh label is the canonical mesh for new resources;
	// the top-level mesh field on the K8s object is often empty on apply.
	mesh := resource.GetMeta().GetMesh()
	if labelMesh, ok := coreLabels[metadata.KumaMeshLabel]; ok && labelMesh != "" {
		mesh = labelMesh
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
