package resources

import (
	// import to register all types
	_ "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

func Global() registry.TypeRegistry {
	return registry.Global()
}
