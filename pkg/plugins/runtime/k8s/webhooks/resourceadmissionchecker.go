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
	if (c.Mode == core.Global || c.FederatedZone) && r.Descriptor().IsPluginOriginated {
		return c.validateLabels(r, ns, op)
	}
	if violations := resource_labels.ValidateOriginFormat(r.GetMeta().GetLabels()); len(violations) > 0 {
		return AdmissionDecision{Response: *forbiddenResponse("Operation not allowed. " + violations[0].String())}
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
	name := r.GetMeta().GetName()
	if k8sName, ok := r.GetMeta().GetNameExtensions()[core_model.K8sNameComponent]; ok && k8sName != "" {
		name = k8sName
	} else if ns != "" {
		name = strings.TrimSuffix(name, "."+ns)
	}
	ctx := resource_labels.ValidationContext{
		Mode:                         c.Mode,
		Env:                          core.KubernetesEnvironment,
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

	if op == v1.Delete {
		if violations := resource_labels.ValidateOrigin(labels, ctx); len(violations) > 0 {
			return AdmissionDecision{Response: *forbiddenResponse("Operation not allowed. " + violations[0].String())}
		}
		return AdmissionDecision{Response: admission.Allowed("")}
	}

	result := resource_labels.Validate(labels, ctx)
	if len(result.Errors) > 0 {
		if result.Errors[0].Format {
			return AdmissionDecision{Response: admission.Allowed("")}
		}
		return AdmissionDecision{Response: *forbiddenResponse("Operation not allowed. " + result.Errors[0].String())}
	}
	warnings := make([]string, 0, len(result.Warnings))
	for _, w := range result.Warnings {
		warnings = append(warnings, w.String())
	}
	return AdmissionDecision{
		Response: admission.Allowed("").WithWarnings(warnings...),
		Warnings: warnings,
	}
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
