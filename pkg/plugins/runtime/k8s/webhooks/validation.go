package webhooks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/validators"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/version"
)

func NewValidatingWebhook(converter k8s_common.Converter, coreRegistry core_registry.TypeRegistry, k8sRegistry k8s_registry.TypeRegistry, mode core.CpMode, cpServiceAccountName string) k8s_common.AdmissionValidator {
	return &validatingHandler{
		coreRegistry:         coreRegistry,
		k8sRegistry:          k8sRegistry,
		converter:            converter,
		mode:                 mode,
		cpServiceAccountName: cpServiceAccountName,
	}
}

type validatingHandler struct {
	coreRegistry         core_registry.TypeRegistry
	k8sRegistry          k8s_registry.TypeRegistry
	converter            k8s_common.Converter
	decoder              *admission.Decoder
	mode                 core.CpMode
	cpServiceAccountName string
}

func (h *validatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *validatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	resType := core_model.ResourceType(req.Kind.Kind)

	_, err := h.coreRegistry.DescriptorFor(resType)
	if err != nil {
		// we only care about types in the registry for this handler
		return admission.Allowed("")
	}

	if resp := h.isOperationAllowed(resType, req.UserInfo, req.Operation); !resp.Allowed {
		return resp
	}

	switch req.Operation {
	case v1.Delete:
		return admission.Allowed("")
	default:
		if resp := h.validateResourceLocation(resType); !resp.Allowed {
			return resp
		}

		coreRes, k8sObj, err := h.decode(req)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if err := core_mesh.ValidateMesh(k8sObj.GetMesh(), coreRes.Descriptor().Scope); err.HasViolations() {
			return convertValidationErrorOf(err, k8sObj, k8sObj.GetObjectMeta())
		}

		if err := coreRes.Validate(); err != nil {
			if kumaErr, ok := err.(*validators.ValidationError); ok {
				// we assume that coreRes.Validate() returns validation errors of the spec
				return convertSpecValidationError(kumaErr, k8sObj)
			}
			return admission.Denied(err.Error())
		}

		return admission.Allowed("")
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
	if err := h.decoder.Decode(req, k8sObj); err != nil {
		return nil, nil, err
	}
	if err := h.converter.ToCoreResource(k8sObj, coreRes); err != nil {
		return nil, nil, err
	}
	return coreRes, k8sObj, nil
}

// Note that this func does not validate ConfigMap and Secret since this webhook does not support those
func (h *validatingHandler) isOperationAllowed(resType core_model.ResourceType, userInfo authenticationv1.UserInfo, op v1.Operation) admission.Response {
	if userInfo.Username == h.cpServiceAccountName {
		// Assume this means sync from another zone. Not security; protecting user from self.
		return admission.Allowed("")
	}

	descriptor, err := h.coreRegistry.DescriptorFor(resType)
	if err != nil {
		return syncErrorResponse(resType, h.mode, op)
	}
	if (h.mode == core.Global && descriptor.KDSFlags.Has(core_model.ConsumedByGlobal)) || (h.mode == core.Zone && resType != core_mesh.DataplaneType && descriptor.KDSFlags.Has(core_model.ConsumedByZone)) {
		return syncErrorResponse(resType, h.mode, op)
	}
	return admission.Allowed("")
}

func syncErrorResponse(resType core_model.ResourceType, cpMode core.CpMode, op v1.Operation) admission.Response {
	otherCpMode := ""
	if cpMode == core.Zone {
		otherCpMode = core.Global
	} else if cpMode == core.Global {
		otherCpMode = core.Zone
	}
	return admission.Response{
		AdmissionResponse: v1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status: "Failure",
				Message: fmt.Sprintf("Operation not allowed. %s resources like %s can be updated or deleted only "+
					"from the %s control plane and not from a %s control plane.", version.Product, resType, strings.ToUpper(otherCpMode), strings.ToUpper(cpMode)),
				Reason: "Forbidden",
				Code:   403,
				Details: &metav1.StatusDetails{
					Causes: []metav1.StatusCause{
						{
							Type:    "FieldValueInvalid",
							Message: "cannot be empty",
							Field:   "metadata.annotations[kuma.io/synced]",
						},
					},
				},
			},
		},
	}
}

// validateResourceLocation validates if resources that suppose to be applied on Global are applied on Global and other way around
func (h *validatingHandler) validateResourceLocation(resType core_model.ResourceType) admission.Response {
	if err := system.ValidateLocation(resType, h.mode); err != nil {
		return admission.Response{
			AdmissionResponse: v1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Status:  "Failure",
					Message: err.Error(),
					Reason:  "Forbidden",
					Code:    403,
				},
			},
		}
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
