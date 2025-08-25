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
	case port == nil:
		return "", errors.New("port is nil: expected a non-nil port implementing core.Port")
	case unifiedNaming:
		return kri.WithSectionName(kri.From(dest), port.GetName()).String(), nil
	default:
		return ResolveLegacyFromDestination(dest, port), nil
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
