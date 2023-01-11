package resources

import (

	// import to register all types
	_ "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	// import to register all types
	_ "github.com/kumahq/kuma/pkg/test/resources/apis/sample"
)

func Global() registry.TypeRegistry {
	return registry.Global()
}
