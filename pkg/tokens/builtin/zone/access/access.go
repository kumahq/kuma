package access

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/user"
)

type ZoneTokenAccess interface {
	ValidateGenerateZoneToken(ctx context.Context, zone string, user user.User) error
}
