package destinationname

import (
	"fmt"

	"github.com/kumahq/kuma/pkg/core/kri"
)

// LegacyName is a current way of naming envoy resources until https://github.com/kumahq/kuma/pull/12756
// is fully implemented
func LegacyName(id kri.Identifier, shortName string, port uint32) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s_%d", id.Mesh, id.Name, id.Namespace, id.Zone, shortName, port)
}
