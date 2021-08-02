package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
)

var GlobalSecretWsDefinition = ResourceWsDefinition{
	Type:  system.GlobalSecretType,
	Path:  "global-secrets",
	Admin: true,
}
