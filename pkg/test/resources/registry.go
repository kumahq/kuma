package resources

import (
	_ "github.com/Kong/kuma/pkg/core/resources/apis/mesh" // import to register all types
	"github.com/Kong/kuma/pkg/core/resources/registry"
	_ "github.com/Kong/kuma/pkg/test/resources/apis/sample" // import to register all types
)

func Global() registry.TypeRegistry {
	return registry.Global()
}
