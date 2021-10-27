package match

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

// ExternalService selects the entries from services that match
// serviceTags. Note that for service matching to work correctly,
// both the service and the tags must have the `kuma.io/service` tag.
func ExternalService(
	services *mesh.ExternalServiceResourceList,
	serviceTags mesh_proto.TagSelector,
) mesh.ExternalServiceResourceList {
	var matched mesh.ExternalServiceResourceList

	for _, s := range services.Items {
		if serviceTags.Matches(s.Spec.GetTags()) {
			if err := matched.AddItem(s); err != nil {
				panic(err.Error()) // Can't fail because we have consistent types.
			}
		}
	}

	return matched
}
