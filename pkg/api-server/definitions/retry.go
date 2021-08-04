package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var RetryWsDefinition = ResourceWsDefinition{
	Type: mesh.RetryType,
	Path: "retries",
}
