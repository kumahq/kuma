package builders_test

import (
	"testing"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
)

func TestName(t *testing.T) {
	out := samples.MeshMTLSBuilder().
		WithName("xyz").
		WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Everywhere).
		WithPermissiveMTLSBackends().UniYaml()
	println(out)
}
