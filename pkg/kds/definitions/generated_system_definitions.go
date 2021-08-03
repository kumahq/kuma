// Generated by tools/resource-gen.
// Run "make generate" to update this file.

// nolint:whitespace
package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
)

var systemDefinitions = []KdsDefinition{

	{
		Type:      system.ConfigType,
		Direction: FromGlobalToZone,
	},

	{
		Type:      system.SecretType,
		Direction: FromGlobalToZone,
	},
}
