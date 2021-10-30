package clusters_test

import (
	"testing"
	"time"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/test"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func DefaultTimeout() *core_mesh.TimeoutResource {
	return &core_mesh.TimeoutResource{Spec: &mesh_proto.Timeout{
		Conf: &mesh_proto.Timeout_Conf{
			ConnectTimeout: util_proto.Duration(5 * time.Second),
		},
	}}
}

func TestClusters(t *testing.T) {
	test.RunSpecs(t, "Clusters V3 Suite")
}
