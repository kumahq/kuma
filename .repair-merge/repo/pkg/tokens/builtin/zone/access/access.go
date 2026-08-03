package access

import (
	"context"

	"github.com/kumahq/kuma/v3/pkg/core/user"
)

type ZoneTokenAccess interface {
	ValidateGenerateZoneToken(ctx context.Context, zone string, user user.User) error
}
