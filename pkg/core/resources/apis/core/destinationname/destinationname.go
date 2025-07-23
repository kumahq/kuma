package destinationname

import (
	"errors"
	"fmt"

	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core"
)

func MustResolve(unifiedNaming bool, dest core.Destination, port core.Port) string {
	name, err := Resolve(unifiedNaming, dest, port)
	if err != nil {
		panic(err)
	}
	return name
}

func Resolve(unifiedNaming bool, dest core.Destination, port core.Port) (string, error) {
	switch {
	case dest == nil:
		return "", errors.New("dest is nil: expected a non-nil dest implementing core.Destination")
	case unifiedNaming && port != nil && port.GetValue() > 0:
		return kri.From(dest, port.GetName()).String(), nil
	case unifiedNaming:
		return kri.From(dest, "").String(), nil
	case port != nil && port.GetValue() > 0:
		return legacyName(dest, port), nil
	default:
		return "", errors.New("destination port is required and must be greater than 0 when unified naming is disabled")
	}
}

// legacyName is a current way of naming envoy resources until https://github.com/kumahq/kuma/pull/12756
// is fully implemented
func legacyName(dest core.Destination, port core.Port) string {
	id := kri.From(dest, "")

	return fmt.Sprintf(
		"%s_%s_%s_%s_%s_%d",
		id.Mesh,
		id.Name,
		id.Namespace,
		id.Zone,
		dest.Descriptor().ShortName,
		port.GetValue(),
	)
}
