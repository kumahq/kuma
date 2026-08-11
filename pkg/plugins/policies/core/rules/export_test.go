package rules

import core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"

// BuildToListWithRoutesForTesting exposes buildToListWithRoutes so tests can
// exercise MeshHTTPRoute expansion (including namespace scoping) directly.
func BuildToListWithRoutesForTesting(meta core_model.ResourceMeta, policyWithTo core_model.PolicyWithToList, httpRoutes []core_model.Resource) ([]core_model.PolicyItem, error) {
	return buildToListWithRoutes(meta, policyWithTo, httpRoutes)
}
