package v1alpha1

import (
	"github.com/kumahq/kuma/v3/pkg/core/resources/extensions"
)

// ExtensionPoint describes `spec.extension` so that products plugging a generator
// into it can have their configuration documented in the OpenAPI spec instead of
// showing up as an opaque `config`. See extensions.Register.
var ExtensionPoint = extensions.Point{
	ResourceType:   HostnameGeneratorType,
	SchemaPath:     []string{"spec", "extension"},
	Discriminator:  "type",
	ConfigProperty: "config",
}
