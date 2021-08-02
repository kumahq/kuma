package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
)

var SecretWsDefinition = ResourceWsDefinition{
	Type:  system.SecretType,
	Path:  "secrets",
	Admin: true,
}
