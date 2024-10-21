package webhooks

import (
	"context"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/validators"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func NewValidatingWebhook(
	converter k8s_common.Converter,
	coreRegistry core_registry.TypeRegistry,
	k8sRegistry k8s_registry.TypeRegistry,
	checker ResourceAdmissionChecker,
) k8s_common.AdmissionValidator {
	return &validatingHandler{
		coreRegistry:             coreRegistry,
		k8sRegistry:              k8sRegistry,
		converter:                converter,
		ResourceAdmissionChecker: checker,
	}
}

type validatingHandler struct {
	ResourceAdmissionChecker

	coreRegistry core_registry.TypeRegistry
	k8sRegistry  k8s_registry.TypeRegistry
	converter    k8s_common.Converter
	decoder      admission.Decoder
}

func (h *validatingHandler) InjectDecoder(d admission.Decoder) {
	h.decoder = d
}

func (h *validatingHandler) Handle(_ context.Context, req admission.Request) admission.Response {
	_, err := h.coreRegistry.DescriptorFor(core_model.ResourceType(req.Kind.Kind))
	if err != nil {
		// we only care about types in the registry for this handler
		return admission.Allowed("")
	}

	coreRes, k8sObj, err := h.decode(req)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if resp := h.IsOperationAllowed(req.UserInfo, coreRes, req.Namespace); !resp.Allowed {
		return resp
	}

	switch req.Operation {
	case v1.Delete:
		return admission.Allowed("")
	default:
		if err := core_mesh.ValidateMesh(k8sObj.GetMesh(), coreRes.Descriptor().Scope); err.HasViolations() {
			return convertValidationErrorOf(err, k8sObj, k8sObj.GetObjectMeta())
		}

		if err := h.validateLabels(coreRes.GetMeta()); err.HasViolations() {
			return convertValidationErrorOf(err, k8sObj, k8sObj.GetObjectMeta())
		}

		if err := core_model.Validate(coreRes); err != nil {
			if kumaErr, ok := err.(*validators.ValidationError); ok {
				// we assume that coreRes.Validate() returns validation errors of the spec
				return convertSpecValidationError(kumaErr, coreRes.Descriptor().IsPluginOriginated, k8sObj)
			}
			return admission.Denied(err.Error())
		}

		return admission.Allowed("").WithWarnings(core_model.Deprecations(coreRes)...)
	}
}

func (h *validatingHandler) decode(req admission.Request) (core_model.Resource, k8s_model.KubernetesObject, error) {
	coreRes, err := h.coreRegistry.NewObject(core_model.ResourceType(req.Kind.Kind))
	if err != nil {
		return nil, nil, err
	}
	k8sObj, err := h.k8sRegistry.NewObject(coreRes.GetSpec())
	if err != nil {
		return nil, nil, err
	}

	switch req.Operation {
	case v1.Delete:
		if err := h.decoder.DecodeRaw(req.OldObject, k8sObj); err != nil {
			return nil, nil, err
		}
	default:
		if err := h.decoder.Decode(req, k8sObj); err != nil {
			return nil, nil, err
		}
	}

	if err := h.converter.ToCoreResource(k8sObj, coreRes); err != nil {
		return nil, nil, err
	}
	return coreRes, k8sObj, nil
}

func (h *validatingHandler) validateLabels(rm core_model.ResourceMeta) validators.ValidationError {
	var verr validators.ValidationError
	labelsPath := validators.Root().Field("labels")
	if origin, ok := core_model.ResourceOrigin(rm); ok {
		if err := origin.IsValid(); err != nil {
			verr.AddViolationAt(labelsPath.Key(mesh_proto.ResourceOriginLabel), err.Error())
		}
	}
	return verr
}

func (h *validatingHandler) Supports(admission.Request) bool {
	return true
}

func convertSpecValidationError(kumaErr *validators.ValidationError, isPluginOriginated bool, obj k8s_model.KubernetesObject) admission.Response {
	verr := validators.OK()
	if kumaErr != nil {
		if isPluginOriginated {
			verr = *kumaErr
		} else {
			verr.AddError("spec", *kumaErr)
		}
	}
	return convertValidationErrorOf(verr, obj, obj.GetObjectMeta())
}

func convertValidationErrorOf(kumaErr validators.ValidationError, obj kube_runtime.Object, objMeta metav1.Object) admission.Response {
	details := &metav1.StatusDetails{
		Name: objMeta.GetName(),
		Kind: obj.GetObjectKind().GroupVersionKind().Kind,
	}
	resp := admission.Response{
		AdmissionResponse: v1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  "Failure",
				Message: kumaErr.Error(),
				Reason:  "Invalid",
				Code:    int32(422),
				Details: details,
			},
		},
	}
	for _, violation := range kumaErr.Violations {
		cause := metav1.StatusCause{
			Type:    "FieldValueInvalid",
			Message: violation.Message,
			Field:   violation.Field,
		}
		details.Causes = append(details.Causes, cause)
	}
	return resp
}
