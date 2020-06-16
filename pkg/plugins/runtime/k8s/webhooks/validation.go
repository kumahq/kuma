package webhooks

import (
	"context"
	"net/http"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_registry "github.com/Kong/kuma/pkg/core/resources/registry"
	"github.com/Kong/kuma/pkg/core/validators"
	k8s_resources "github.com/Kong/kuma/pkg/plugins/resources/k8s"
	k8s_model "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

func NewValidatingWebhook(converter k8s_resources.Converter, coreRegistry core_registry.TypeRegistry, k8sRegistry k8s_registry.TypeRegistry) AdmissionValidator {
	return &validatingHandler{
		coreRegistry: coreRegistry,
		k8sRegistry:  k8sRegistry,
		converter:    converter,
	}
}

type validatingHandler struct {
	coreRegistry core_registry.TypeRegistry
	k8sRegistry  k8s_registry.TypeRegistry
	converter    k8s_resources.Converter
	decoder      *admission.Decoder
}

func (h *validatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *validatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation == v1beta1.Delete {
		return admission.Allowed("")
	}

	resType := core_model.ResourceType(req.Kind.Kind)

	coreRes, err := h.coreRegistry.NewObject(resType)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	obj, err := h.k8sRegistry.NewObject(coreRes.GetSpec())
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	// unmarshal k8s object from the request
	if err := h.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := h.converter.ToCoreResource(obj.(k8s_model.KubernetesObject), coreRes); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if err := coreRes.Validate(); err != nil {
		if kumaErr, ok := err.(*validators.ValidationError); ok {
			// we assume that coreRes.Validate() returns validation errors of the spec
			return convertSpecValidationError(kumaErr, obj)
		}
		return admission.Denied(err.Error())
	}

	return admission.Allowed("")
}

func (h *validatingHandler) Supports(admission.Request) bool {
	return true
}

func convertSpecValidationError(kumaErr *validators.ValidationError, obj k8s_model.KubernetesObject) admission.Response {
	verr := validators.ValidationError{}
	if kumaErr != nil {
		verr.AddError("spec", *kumaErr)
	}
	return convertValidationErrorOf(verr, obj, obj.GetObjectMeta())
}

func convertValidationErrorOf(kumaErr validators.ValidationError, obj kube_runtime.Object, objMeta metav1.Object) admission.Response {
	details := &metav1.StatusDetails{
		Name: objMeta.GetName(),
		Kind: obj.GetObjectKind().GroupVersionKind().Kind,
	}
	resp := admission.Response{
		AdmissionResponse: v1beta1.AdmissionResponse{
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
