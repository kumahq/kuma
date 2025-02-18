package schema

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis"
	"github.com/kumahq/kuma/pkg/plugins/policies"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

func IsClusterScopeResource(resourceType string) bool {
	if isClusterScope, exist := v1alpha1.ResourceScope[resourceType]; exist {
		return isClusterScope
	}
	if isClusterScope, exist := policies.ResourceToScope[resourceType]; exist {
		return isClusterScope
	}
	if isClusterScope, exist := apis.ResourceToScope[resourceType]; exist {
		return isClusterScope
	}
	return false
}
