package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
<<<<<<< HEAD
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
=======
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
>>>>>>> da824ce57 (fix(kuma-cp): mistakenly setting 'kuma.io/display-name' as label (#10430))
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

func (h *defaultingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
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

	if resp := h.IsOperationAllowed(req.UserInfo, resource); !resp.Allowed {
		return resp
	}
<<<<<<< HEAD

	if resource.Descriptor().Scope == core_model.ScopeMesh {
		labels := obj.GetLabels()
		if _, ok := labels[metadata.KumaMeshLabel]; !ok {
			if len(labels) == 0 {
				labels = map[string]string{}
			}
			labels[metadata.KumaMeshLabel] = core_model.DefaultMesh
			obj.SetLabels(labels)
		}
	}

	if h.Mode == core.Zone {
		labels := obj.GetLabels()
		if _, ok := core_model.ResourceOrigin(resource.GetMeta()); !ok {
			if len(labels) == 0 {
				labels = map[string]string{}
			}
			labels[mesh_proto.ResourceOriginLabel] = string(mesh_proto.ZoneResourceOrigin)
			obj.SetLabels(labels)
		}
	}
=======
	labels, annotations := k8s.SplitLabelsAndAnnotations(
		core_model.ComputeLabels(resource, h.Mode, true, h.SystemNamespace),
		obj.GetAnnotations(),
	)
	obj.SetLabels(labels)
	obj.SetAnnotations(annotations)
>>>>>>> da824ce57 (fix(kuma-cp): mistakenly setting 'kuma.io/display-name' as label (#10430))

	marshaled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}
