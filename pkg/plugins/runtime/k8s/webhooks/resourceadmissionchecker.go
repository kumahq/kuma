package webhooks

import (
	"fmt"
	"slices"
	"strings"

	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kumahq/kuma/v3/pkg/config/core"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/version"
)

type ResourceAdmissionChecker struct {
	AllowedUsers    []string
	ControlPlane    resource_labels.ControlPlane
	SystemNamespace string
}

const (
	GenericGarbageCollectorUser = "system:serviceaccount:kube-system:generic-garbage-collector"
	StorageVersionMigratorUser  = "system:serviceaccount:kube-system:storage-version-migrator-controller"
)

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

	if err := resource_labels.ValidateOwnership(resource_labels.Write{
		Descriptor:  r.Descriptor(),
		Spec:        r.GetSpec(),
		Namespace:   resource_labels.NewNamespace(ns, ns == c.SystemNamespace),
		Mesh:        r.GetMeta().GetMesh(),
		DisplayName: r.GetMeta().GetName(),
		Labels:      r.GetMeta().GetLabels(),
	}, c.ControlPlane); err.HasViolations() {
		return *forbiddenResponse("Operation not allowed. " + err.Violations[0].Message)
	}

	if r.Descriptor().IsReadOnly(c.ControlPlane.Mode == core.Global, c.ControlPlane.FederatedZone) {
		return *forbiddenResponse(resourceTypeNotAllowedMsg(r.Descriptor().Name, c.ControlPlane.Mode))
	}

	return admission.Allowed("")
}

func (c *ResourceAdmissionChecker) isNamespaceAllowed(r core_model.Resource, ns string) admission.Response {
	switch c.ControlPlane.Mode {
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

func (c *ResourceAdmissionChecker) isPrivilegedUser(allowedUsers []string, userInfo authenticationv1.UserInfo) bool {
	// Assume this means one of the following:
	// - sync from another zone
	// - GC cleanup resources due to OwnerRef.
	// - storage-version migration
	// Not security; protecting user from self.
	return slices.Contains(allowedUsers, userInfo.Username)
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
		Allowed: false,
		Result: &metav1.Status{
			Status:  "Failure",
			Message: msg,
			Reason:  "Forbidden",
			Code:    403,
		},
	}
}
