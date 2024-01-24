package samples

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

func Retry() *core_mesh.RetryResource {
	return builders.Retry().Build()
}
