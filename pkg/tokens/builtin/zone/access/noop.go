package access

import (
	"context"

	"github.com/kumahq/kuma/v2/pkg/core/user"
	"github.com/kumahq/kuma/v2/pkg/tokens/builtin/zone"
)

type NoopZoneTokenAccess struct{}

var _ ZoneTokenAccess = NoopZoneTokenAccess{}

func (n NoopZoneTokenAccess) ValidateGenerateZoneToken(ctx context.Context, identity zone.Identity, user user.User) error {
	return nil
}
