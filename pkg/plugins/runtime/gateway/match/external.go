package match

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

// ExternalService selects the entries from services that match
// serviceTags. Note that for service matching to work correctly,
// both the service and the tags must have the `kuma.io/service` tag.
func ExternalService(
	services []*mesh.ExternalServiceResource,
	serviceTags mesh_proto.TagSelector,
) []*mesh.ExternalServiceResource {
	var matched []*mesh.ExternalServiceResource

	for _, s := range services {
		if serviceTags.Matches(s.Spec.GetTags()) {
			matched = append(matched, s)
		}
	}

	return matched
}
