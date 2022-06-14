package clusters_test

import (
	"testing"
	"time"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/test"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func DefaultTimeout() *mesh_proto.Timeout_Conf {
	return &mesh_proto.Timeout_Conf{
		ConnectTimeout: util_proto.Duration(5 * time.Second),
	}
}

func TestClusters(t *testing.T) {
	test.RunSpecs(t, "Clusters V3 Suite")
}
