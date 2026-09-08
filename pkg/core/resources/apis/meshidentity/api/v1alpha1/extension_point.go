package v1alpha1

import (
	"github.com/kumahq/kuma/v3/pkg/core/resources/extensions"
)

// ExtensionPoint describes `spec.provider.extension`, reached when the provider
// type is Extension, so that products plugging an identity provider into it can
// have their configuration documented in the OpenAPI spec instead of showing up
// as an opaque `config`. See extensions.Register.
var ExtensionPoint = extensions.Point{
	ResourceType:  MeshIdentityType,
	SchemaPath:    []string{"spec", "provider", "extension"},
	Discriminator: "name",
}
