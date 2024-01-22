package webhooks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/exp/slices"
	v1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/validators"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/version"
)

func NewValidatingWebhook(
	converter k8s_common.Converter,
	coreRegistry core_registry.TypeRegistry,
	k8sRegistry k8s_registry.TypeRegistry,
	mode core.CpMode,
	federatedZone bool,
	allowedUsers []string,
	disableOriginLabelValidation bool,
) k8s_common.AdmissionValidator {
	return &validatingHandler{
		coreRegistry:                 coreRegistry,
		k8sRegistry:                  k8sRegistry,
		converter:                    converter,
		mode:                         mode,
		federatedZone:                federatedZone,
		allowedUsers:                 allowedUsers,
		disableOriginLabelValidation: disableOriginLabelValidation,
	}
}

type validatingHandler struct {
	coreRegistry                 core_registry.TypeRegistry
	k8sRegistry                  k8s_registry.TypeRegistry
	converter                    k8s_common.Converter
	decoder                      *admission.Decoder
	mode                         core.CpMode
	federatedZone                bool
	allowedUsers                 []string
	disableOriginLabelValidation bool
}

func (h *validatingHandler) InjectDecoder(d *admission.Decoder) {
	h.decoder = d
}

func (h *validatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	_, err := h.coreRegistry.DescriptorFor(core_model.ResourceType(req.Kind.Kind))
	if err != nil {
		// we only care about types in the registry for this handler
		return admission.Allowed("")
	}

	coreRes, k8sObj, err := h.decode(req)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if resp := h.isOperationAllowed(req.UserInfo, coreRes); !resp.Allowed {
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
				return convertSpecValidationError(kumaErr, coreRes.Descriptor().IsTargetRefBased, k8sObj)
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

// Note that this func does not validate ConfigMap and Secret since this webhook does not support those
func (h *validatingHandler) isOperationAllowed(userInfo authenticationv1.UserInfo, r core_model.Resource) admission.Response {
	if h.isPrivilegedUser(userInfo) {
		return admission.Allowed("")
	}

	if !h.isResourceTypeAllowed(r.Descriptor()) {
		return resourceTypeIsNotAllowedResponse(r.Descriptor().Name, h.mode)
	}

	if !h.isResourceAllowed(r) {
		return resourceIsNotAllowedResponse()
	}

	return admission.Allowed("")
}

func (h *validatingHandler) isPrivilegedUser(userInfo authenticationv1.UserInfo) bool {
	// Assume this means one of the following:
	// - sync from another zone (rt.Config().Runtime.Kubernetes.ServiceAccountName)
	// - GC cleanup resources due to OwnerRef. ("system:serviceaccount:kube-system:generic-garbage-collector")
	// - storageversionmigratior
	// Not security; protecting user from self.
	return slices.Contains(h.allowedUsers, userInfo.Username)
}

func (h *validatingHandler) isResourceTypeAllowed(d core_model.ResourceTypeDescriptor) bool {
	if d.KDSFlags == core_model.KDSDisabledFlag {
		return true
	}
	if h.mode == core.Global && !d.KDSFlags.Has(core_model.AllowedOnGlobalSelector) {
		return false
	}
	if h.federatedZone && !d.KDSFlags.Has(core_model.AllowedOnZoneSelector) {
		return false
	}
	return true
}

func (h *validatingHandler) isResourceAllowed(r core_model.Resource) bool {
	if !h.federatedZone || !r.Descriptor().IsPluginOriginated {
		return true
	}
	if !h.disableOriginLabelValidation {
		if origin, ok := core_model.ResourceOrigin(r.GetMeta()); !ok || origin != mesh_proto.ZoneResourceOrigin {
			return false
		}
	}
	return true
}

func (h *validatingHandler) validateLabels(rm core_model.ResourceMeta) validators.ValidationError {
	var verr validators.ValidationError
	if origin, ok := core_model.ResourceOrigin(rm); ok {
		if err := origin.IsValid(); err != nil {
			verr.AddViolationAt(validators.Root().Field("labels").Key(mesh_proto.ResourceOriginLabel), err.Error())
		}
	}
	return verr
}

func resourceIsNotAllowedResponse() admission.Response {
	return admission.Response{
		AdmissionResponse: v1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  "Failure",
				Message: fmt.Sprintf("Operation not allowed. Applying policies on Zone CP requires '%s' label to be set to '%s'.", mesh_proto.ResourceOriginLabel, mesh_proto.ZoneResourceOrigin),
				Reason:  "Forbidden",
				Code:    403,
				Details: &metav1.StatusDetails{
					Causes: []metav1.StatusCause{
						{
							Type:    "FieldValueInvalid",
							Message: "cannot be empty",
							Field:   "metadata.labels[kuma.io/origin]",
						},
					},
				},
			},
		},
	}
}

func resourceTypeIsNotAllowedResponse(resType core_model.ResourceType, cpMode core.CpMode) admission.Response {
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

func (h *validatingHandler) Supports(admission.Request) bool {
	return true
}

func convertSpecValidationError(kumaErr *validators.ValidationError, isTargetRef bool, obj k8s_model.KubernetesObject) admission.Response {
	verr := validators.OK()
	if kumaErr != nil {
		if isTargetRef {
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
