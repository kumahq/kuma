package system

import (
	"fmt"
	"strings"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

// ValidateLocation checks if System Resources are applied on a proper Control Plane
// Note: ideally, this should be hooked to ResourceManager but for now on Universal we mark API as readonly
// therefore we don't need this validation yet.
func ValidateLocation(resourceType model.ResourceType, mode core.CpMode) error {
	switch resourceType {
	case ZoneType:
		if mode != core.Global {
			return InvalidLocationErr(resourceType, core.Global)
		}
	}
	return nil
}

func InvalidLocationErr(resourceType model.ResourceType, supportedLocations ...core.CpMode) error {
	return fmt.Errorf("%s resource can only be applied on CP with mode: %v", resourceType, supportedLocations)
}

func IsInvalidLocationErr(err error) bool {
	return strings.Contains(err.Error(), "resource can only be applied on CP with mode")
}
