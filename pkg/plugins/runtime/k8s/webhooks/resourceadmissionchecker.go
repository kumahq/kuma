package webhooks

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
	v1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/version"
)

type ResourceAdmissionChecker struct {
	AllowedUsers                 []string
	Mode                         core.CpMode
	FederatedZone                bool
	DisableOriginLabelValidation bool
}

func (c *ResourceAdmissionChecker) IsOperationAllowed(userInfo authenticationv1.UserInfo, r core_model.Resource) admission.Response {
	if c.isPrivilegedUser(c.AllowedUsers, userInfo) {
		return admission.Allowed("")
	}

	if !c.isResourceTypeAllowed(r.Descriptor()) {
		return c.resourceTypeIsNotAllowedResponse(r.Descriptor().Name)
	}

	if !c.isResourceAllowed(r) {
		return c.resourceIsNotAllowedResponse()
	}

	return admission.Allowed("")
}

func (c *ResourceAdmissionChecker) isResourceTypeAllowed(d core_model.ResourceTypeDescriptor) bool {
	if d.KDSFlags == core_model.KDSDisabledFlag {
		return true
	}
	if c.Mode == core.Global && !d.KDSFlags.Has(core_model.AllowedOnGlobalSelector) {
		return false
	}
	if c.FederatedZone && !d.KDSFlags.Has(core_model.AllowedOnZoneSelector) {
		return false
	}
	return true
}

func (c *ResourceAdmissionChecker) isResourceAllowed(r core_model.Resource) bool {
	if !c.FederatedZone || !r.Descriptor().IsPluginOriginated {
		return true
	}
	if !c.DisableOriginLabelValidation {
		if origin, ok := core_model.ResourceOrigin(r.GetMeta()); !ok || origin != mesh_proto.ZoneResourceOrigin {
			return false
		}
	}
	return true
}

func (c *ResourceAdmissionChecker) isPrivilegedUser(allowedUsers []string, userInfo authenticationv1.UserInfo) bool {
	// Assume this means one of the following:
	// - sync from another zone (rt.Config().Runtime.Kubernetes.ServiceAccountName)
	// - GC cleanup resources due to OwnerRef. ("system:serviceaccount:kube-system:generic-garbage-collector")
	// - storageversionmigratior
	// Not security; protecting user from self.
	return slices.Contains(allowedUsers, userInfo.Username)
}

func (c *ResourceAdmissionChecker) resourceIsNotAllowedResponse() admission.Response {
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

func (c *ResourceAdmissionChecker) resourceTypeIsNotAllowedResponse(resType core_model.ResourceType) admission.Response {
	otherCpMode := ""
	if c.Mode == core.Zone {
		otherCpMode = core.Global
	} else if c.Mode == core.Global {
		otherCpMode = core.Zone
	}
	return admission.Response{
		AdmissionResponse: v1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status: "Failure",
				Message: fmt.Sprintf("Operation not allowed. %s resources like %s can be updated or deleted only "+
					"from the %s control plane and not from a %s control plane.", version.Product, resType, strings.ToUpper(otherCpMode), strings.ToUpper(c.Mode)),
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
