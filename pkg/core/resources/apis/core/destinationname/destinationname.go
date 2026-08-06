package destinationname

import (
	"errors"
	"fmt"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core"
)

func MustResolve(dest core.Destination, port core.Port) string {
	name, err := Resolve(dest, port)
	if err != nil {
		panic(err)
	}
	return name
}

// Resolve returns the KRI of a destination port.
func Resolve(dest core.Destination, port core.Port) (string, error) {
	switch {
	case dest == nil:
		return "", errors.New("dest is nil: expected a non-nil dest implementing core.Destination")
	case port == nil:
		return "", errors.New("port is nil: expected a non-nil port implementing core.Port")
	default:
		return kri.WithSectionName(kri.From(dest), port.GetName()).String(), nil
	}
}

func ResolveLegacyFromDestination(dest core.Destination, port core.Port) string {
	return ResolveLegacyFromKRI(
		kri.From(dest),
		dest.Descriptor().ShortName,
		port.GetValue(),
	)
}

func ResolveLegacyFromKRI(id kri.Identifier, resourceTypeShortName string, port int32) string {
	return fmt.Sprintf(
		"%s_%s_%s_%s_%s_%d",
		id.Mesh,
		id.Name,
		id.Namespace,
		id.Zone,
		resourceTypeShortName,
		port,
	)
}
