package resources

import (
	_ "github.com/kumahq/kuma/pkg/core/resources/apis/mesh" // import to register all types
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	_ "github.com/kumahq/kuma/pkg/test/resources/apis/sample" // import to register all types
)

func Global() registry.TypeRegistry {
	return registry.Global()
}
