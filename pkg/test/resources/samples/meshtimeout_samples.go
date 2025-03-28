package samples

import (
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	meshtimeout_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

func MeshTimeoutInCustomNamespace() *meshtimeout_proto.MeshTimeoutResource {
	return builders.MeshTimeout().WithTargetRef(
		builders.TargetRefMesh(),
	).AddTo(
		builders.TargetRefMesh(),
		meshtimeout_proto.Conf{
			IdleTimeout: &v1.Duration{
				Duration: 99 * time.Second,
			},
		},
	).WithNamespace("kuma-test").WithName("mt-in-namespace-kuma-test").Build()
}
