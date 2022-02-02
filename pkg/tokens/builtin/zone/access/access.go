package access

import (
	"github.com/kumahq/kuma/pkg/core/user"
)

type ZoneTokenAccess interface {
	ValidateGenerateZoneToken(zone string, user user.User) error
}
