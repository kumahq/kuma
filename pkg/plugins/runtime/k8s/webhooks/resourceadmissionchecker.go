package webhooks

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	v1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/config/core"
	resource_labels "github.com/kumahq/kuma/v2/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	k8s_resources "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s"
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

func (c *ResourceAdmissionChecker) IsOperationAllowed(userInfo authenticationv1.UserInfo, r core_model.Resource, ns string) admission.Response {
	if c.isPrivilegedUser(c.AllowedUsers, userInfo) {
		return admission.Allowed("")
	}

	if ns != "" {
		// check only namespace-scoped resources
		if resp := c.isNamespaceAllowed(r, ns); !resp.Allowed {
			return resp
		}
	}

	if r.Descriptor().IsReadOnly(c.Mode == core.Global, c.FederatedZone) {
		return *forbiddenResponse(resourceTypeNotAllowedMsg(r.Descriptor().Name, c.Mode))
	}

	if errResponse := c.isResourceAllowed(r, ns); errResponse != nil {
		return *errResponse
	}

	return admission.Allowed("")
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

func (c *ResourceAdmissionChecker) isResourceAllowed(r core_model.Resource, ns string) *admission.Response {
	// we don't need to validate non fedarated zone and legacy policies
	if (c.Mode != core.Global && !c.FederatedZone) || !r.Descriptor().IsPluginOriginated {
		return nil
	}
	return c.validateLabels(r, ns)
}

func (c *ResourceAdmissionChecker) isPrivilegedUser(allowedUsers []string, userInfo authenticationv1.UserInfo) bool {
	// Assume this means one of the following:
	// - sync from another zone
	// - GC cleanup resources due to OwnerRef. ("system:serviceaccount:kube-system:generic-garbage-collector")
	// - storageversionmigratior
	// Not security; protecting user from self.
	return slices.Contains(allowedUsers, userInfo.Username)
}

func (c *ResourceAdmissionChecker) validateLabels(r core_model.Resource, ns string) *admission.Response {
	// On K8s the kuma core name is "<metadata.name>.<namespace>". For label
	// validation (notably kuma.io/display-name, which equals the K8s
	// metadata.name) we need the unqualified name, so strip the namespace
	// suffix.
	name := r.GetMeta().GetName()
	if ns != "" {
		name = strings.TrimSuffix(name, "."+ns)
	}
	ctx := resource_labels.ValidationContext{
		Mode:                         c.Mode,
		IsK8s:                        true,
		FederatedZone:                c.FederatedZone,
		ZoneName:                     c.ZoneName,
		SystemNamespace:              c.SystemNamespace,
		Namespace:                    resource_labels.NewNamespace(ns, ns == c.SystemNamespace),
		DisableOriginLabelValidation: c.DisableOriginLabelValidation,
		Descriptor:                   r.Descriptor(),
		Spec:                         r.GetSpec(),
		ResourceName:                 name,
		ResourceMesh:                 r.GetMeta().GetMesh(),
	}
	violations := resource_labels.Validate(rawK8sLabels(r.GetMeta()), ctx)
	if len(violations) == 0 {
		return nil
	}
	// Defer to the K8s API server's native "Invalid value: ..." error when the
	// problem is label format — its message carries kind/name/key context that
	// our admission response doesn't.
	if violations[0].Format {
		return nil
	}
	return forbiddenResponse("Operation not allowed. " + formatLabelViolation(violations[0]))
}

// rawK8sLabels returns what the user actually supplied on the K8s object,
// before KubernetesMetaAdapter.GetLabels rewrites kuma.io/display-name (and
// similar) from the metadata.name fallback. Reserved-namespace annotations
// (the canonical storage location for kuma.io/display-name) are folded in so
// the validator observes them too.
func rawK8sLabels(meta core_model.ResourceMeta) map[string]string {
	adapter, ok := meta.(*k8s_resources.KubernetesMetaAdapter)
	if !ok {
		return meta.GetLabels()
	}
	labels := maps.Clone(adapter.Labels)
	if labels == nil {
		labels = map[string]string{}
	}
	for k, v := range adapter.Annotations {
		if !mesh_proto.IsReservedLabelKey(k) {
			continue
		}
		if _, present := labels[k]; present {
			continue
		}
		labels[k] = v
	}
	return labels
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
