package webhooks

import (
	"fmt"
	"slices"
	"strings"

	v1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kumahq/kuma/v2/pkg/config/core"
	resource_labels "github.com/kumahq/kuma/v2/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/version"
)

type ResourceAdmissionChecker struct {
	AllowedUsers                 []string
	Mode                         core.CpMode
	FederatedZone                bool
	DisableOriginLabelValidation bool
	SystemNamespace              string
	ZoneName                     string
}

// AdmissionDecision is the result of an admission check. When Response.Allowed
// is false, Warnings is empty. When Allowed is true, Warnings carries any
// non-blocking findings (e.g. user-supplied CP-managed label values that the
// CP will override on persist via Compute).
//
// Privileged signals that the request bypassed validation (KDS sync, GC,
// storage-version migrator). The mutating defaulter must NOT run Compute on
// these — the resource's labels are authoritative from the source CP and
// recomputing would clobber kuma.io/origin, kuma.io/zone, etc.
type AdmissionDecision struct {
	Response   admission.Response
	Warnings   []string
	Privileged bool
}

func (c *ResourceAdmissionChecker) IsOperationAllowed(userInfo authenticationv1.UserInfo, r core_model.Resource, ns string, op v1.Operation) AdmissionDecision {
	if c.isPrivilegedUser(c.AllowedUsers, userInfo) {
		return AdmissionDecision{Response: admission.Allowed(""), Privileged: true}
	}

	if ns != "" {
		// check only namespace-scoped resources
		if resp := c.isNamespaceAllowed(r, ns); !resp.Allowed {
			return AdmissionDecision{Response: resp}
		}
	}

	if r.Descriptor().IsReadOnly(c.Mode == core.Global, c.FederatedZone) {
		return AdmissionDecision{Response: *forbiddenResponse(resourceTypeNotAllowedMsg(r.Descriptor().Name, c.Mode))}
	}

	return c.checkResource(r, ns, op)
}

func (c *ResourceAdmissionChecker) isNamespaceAllowed(r core_model.Resource, ns string) admission.Response {
	switch c.Mode {
	case core.Global:
		if ns != c.SystemNamespace {
			return admission.Denied(fmt.Sprintf("on Global CP the policy can be created only in the system namespace:%s", c.SystemNamespace))
		}
	case core.Zone:
		if r.Descriptor().AllowedOnSystemNamespaceOnly && ns != c.SystemNamespace {
			return admission.Denied(fmt.Sprintf("resource type %v can be created only in the system namespace:%s", r.Descriptor().Name, c.SystemNamespace))
		}
	}
	return admission.Allowed("")
}

func (c *ResourceAdmissionChecker) checkResource(r core_model.Resource, ns string, op v1.Operation) AdmissionDecision {
	// Federated/plugin resources get the full context-aware validation, which
	// covers both the origin vocabulary check and CP-owned label mismatches
	// (now surfaced as warnings, not errors). Non-federated zones / legacy
	// resources skip that broader check but still get the origin vocabulary
	// check below.
	if (c.Mode == core.Global || c.FederatedZone) && r.Descriptor().IsPluginOriginated {
		return c.validateLabels(r, ns, op)
	}
	if violations := resource_labels.ValidateOriginFormat(r.GetMeta().GetLabels()); len(violations) > 0 {
		return AdmissionDecision{Response: *forbiddenResponse("Operation not allowed. " + formatLabelViolation(violations[0]))}
	}
	return AdmissionDecision{Response: admission.Allowed("")}
}

func (c *ResourceAdmissionChecker) isPrivilegedUser(allowedUsers []string, userInfo authenticationv1.UserInfo) bool {
	// Assume this means one of the following:
	// - sync from another zone
	// - GC cleanup resources due to OwnerRef. ("system:serviceaccount:kube-system:generic-garbage-collector")
	// - storageversionmigratior
	// Not security; protecting user from self.
	return slices.Contains(allowedUsers, userInfo.Username)
}

func (c *ResourceAdmissionChecker) validateLabels(r core_model.Resource, ns string, op v1.Operation) AdmissionDecision {
	// On K8s the kuma core name is "<metadata.name>.<namespace>". For label
	// validation (notably kuma.io/display-name, which equals the K8s
	// metadata.name) we need the unqualified name. GetNameExtensions carries
	// the original K8s metadata.name; fall back to a suffix strip if the
	// extension is missing.
	name := r.GetMeta().GetName()
	if k8sName, ok := r.GetMeta().GetNameExtensions()[core_model.K8sNameComponent]; ok && k8sName != "" {
		name = k8sName
	} else if ns != "" {
		name = strings.TrimSuffix(name, "."+ns)
	}
	ctx := resource_labels.ValidationContext{
		Mode:                         c.Mode,
		IsK8s:                        true,
		FederatedZone:                c.FederatedZone,
		ZoneName:                     c.ZoneName,
		Namespace:                    resource_labels.NewNamespace(ns, ns == c.SystemNamespace),
		DisableOriginLabelValidation: c.DisableOriginLabelValidation,
		Descriptor:                   r.Descriptor(),
		Spec:                         r.GetSpec(),
		ResourceName:                 name,
		ResourceMesh:                 r.GetMeta().GetMesh(),
	}
	labels := r.GetMeta().GetLabels()

	// Delete authorization only hinges on origin — was this resource created
	// in the CP that's now being asked to delete it? Other CP-computed labels
	// reflect their originating CP's context (notably for KDS-synced resources)
	// and aren't meaningful to recompute here. Origin mismatches on delete
	// stay as errors so we don't silently delete from the wrong CP.
	if op == v1.Delete {
		if violations := resource_labels.ValidateOrigin(labels, ctx); len(violations) > 0 {
			return AdmissionDecision{Response: *forbiddenResponse("Operation not allowed. " + formatLabelViolation(violations[0]))}
		}
		return AdmissionDecision{Response: admission.Allowed("")}
	}

	result := resource_labels.Validate(labels, ctx)
	if len(result.Errors) > 0 {
		// Defer to the K8s API server's native "Invalid value: ..." error when
		// the problem is label format — its message carries kind/name/key
		// context that our admission response doesn't.
		if result.Errors[0].Format {
			return AdmissionDecision{Response: admission.Allowed("")}
		}
		return AdmissionDecision{Response: *forbiddenResponse("Operation not allowed. " + formatLabelViolation(result.Errors[0]))}
	}
	warnings := make([]string, 0, len(result.Warnings))
	for _, w := range result.Warnings {
		warnings = append(warnings, formatLabelViolation(w))
	}
	return AdmissionDecision{
		Response: admission.Allowed("").WithWarnings(warnings...),
		Warnings: warnings,
	}
}

// formatLabelViolation renders a Violation into the inline "Operation not
// allowed. <msg>" form used by the K8s admission response. Violation.Reason
// usually already mentions the label key (e.g. "kuma.io/zone should be ...");
// for reasons that don't (the generic "is a reserved label" case) we prefix
// the quoted key so the message remains self-contained.
func formatLabelViolation(v resource_labels.Violation) string {
	if strings.Contains(v.Reason, v.Key) {
		return v.Reason
	}
	return "'" + v.Key + "' " + v.Reason
}

func resourceTypeNotAllowedMsg(resType core_model.ResourceType, mode core.CpMode) string {
	otherCpMode := ""
	switch mode {
	case core.Zone:
		otherCpMode = core.Global
	case core.Global:
		otherCpMode = core.Zone
	}
	return fmt.Sprintf("Operation not allowed. %s resources like %s can be updated or deleted only "+
		"from the %s control plane and not from a %s control plane.", version.Product, resType, strings.ToUpper(otherCpMode), strings.ToUpper(mode))
}

func forbiddenResponse(msg string) *admission.Response {
	return &admission.Response{
		AdmissionResponse: v1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  "Failure",
				Message: msg,
				Reason:  "Forbidden",
				Code:    403,
			},
		},
	}
}
