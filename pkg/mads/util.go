package mads

import (
	"fmt"

	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
)

func DataplaneAssignmentName(dataplane *core_mesh.DataplaneResource) string {
	// unique name, e.g. REST API uri
	return fmt.Sprintf("/meshes/%s/dataplanes/%s", dataplane.Meta.GetMesh(), dataplane.Meta.GetName())
}
