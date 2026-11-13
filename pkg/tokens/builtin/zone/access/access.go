package access

import (
	"context"

	"github.com/kumahq/kuma/v3/pkg/core/user"
	"github.com/kumahq/kuma/v3/pkg/tokens/builtin/zone"
)

type ZoneTokenAccess interface {
	ValidateGenerateZoneToken(ctx context.Context, identity zone.Identity, user user.User) error
}
