package webhooks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"k8s.io/api/admission/v1beta1"
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
)

func NewValidatingWebhook(converter k8s_common.Converter, coreRegistry core_registry.TypeRegistry, k8sRegistry k8s_registry.TypeRegistry, mode core.CpMode, systemNamespace string) k8s_common.AdmissionValidator {
	return &validatingHandler{
		coreRegistry:    coreRegistry,
		k8sRegistry:     k8sRegistry,
		converter:       converter,
		mode:            mode,
		systemNamespace: systemNamespace,
	}
}

type validatingHandler struct {
	coreRegistry    core_registry.TypeRegistry
	k8sRegistry     k8s_registry.TypeRegistry
	converter       k8s_common.Converter
	decoder         *admission.Decoder
	mode            core.CpMode
	systemNamespace string
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
	if resp := h.validateSync(resType, obj, req.UserInfo); !resp.Allowed {
		return resp
	}
	if resp := h.validateResourceLocation(resType, obj); !resp.Allowed {
		return resp
	}

	if err := h.converter.ToCoreResource(obj, coreRes); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if err := core_mesh.ValidateMesh(obj.GetMesh(), coreRes.Descriptor().Scope); err.HasViolations() {
		return convertValidationErrorOf(err, obj, obj.GetObjectMeta())
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

// Note that this func does not validate ConfigMap and Secret since this webhook does not support those
func (h *validatingHandler) validateSync(resType core_model.ResourceType, obj k8s_model.KubernetesObject, userInfo authenticationv1.UserInfo) admission.Response {
	if isDefaultMesh(resType, obj) { // skip validation for the default mesh
		return admission.Allowed("")
	}

	if isKumaServiceAccount(userInfo, h.systemNamespace) {
		// Assume this means sync from another zone. Not security; protecting user from self.
		return admission.Allowed("")
	}

	descriptor, err := h.coreRegistry.DescriptorFor(resType)
	if err != nil {
		return syncErrorResponse(resType, h.mode)
	}
	if (h.mode == core.Global && descriptor.KDSFlags.Has(core_model.ConsumedByGlobal)) || (h.mode == core.Zone && resType != core_mesh.DataplaneType && descriptor.KDSFlags.Has(core_model.ConsumedByZone)) {
		return syncErrorResponse(resType, h.mode)
	}
	return admission.Allowed("")
}

func syncErrorResponse(resType core_model.ResourceType, cpMode core.CpMode) admission.Response {
	otherCpMode := ""
	if cpMode == core.Zone {
		otherCpMode = core.Global
	} else if cpMode == core.Global {
		otherCpMode = core.Zone
	}
	return admission.Response{
		AdmissionResponse: v1beta1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  "Failure",
				Message: fmt.Sprintf("You are trying to apply a %s on %s CP. In multizone setup, it should be only applied on %s CP and synced to %s CP.", resType, cpMode, otherCpMode, cpMode),
				Reason:  "Forbidden",
				Code:    403,
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

func isKumaServiceAccount(userInfo authenticationv1.UserInfo, systemNamespace string) bool {
	elms := strings.Split(userInfo.Username, ":")
	// system:serviceaccount:<namespace>:kuma-control-plane
	if len(elms) == 4 && elms[2] == systemNamespace {
		return true
	}
	return false
}

func isDefaultMesh(resType core_model.ResourceType, obj k8s_model.KubernetesObject) bool {
	return resType == core_mesh.MeshType && obj.GetName() == core_model.DefaultMesh && len(obj.GetSpec()) == 0
}

// validateResourceLocation validates if resources that suppose to be applied on Global are applied on Global and other way around
func (h *validatingHandler) validateResourceLocation(resType core_model.ResourceType, obj k8s_model.KubernetesObject) admission.Response {
	if err := system.ValidateLocation(resType, h.mode); err != nil {
		return admission.Response{
			AdmissionResponse: v1beta1.AdmissionResponse{
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
