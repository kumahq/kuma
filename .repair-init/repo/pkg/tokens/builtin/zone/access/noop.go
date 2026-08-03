package access

import (
	"context"

	"github.com/kumahq/kuma/v3/pkg/core/user"
)

type NoopZoneTokenAccess struct{}

var _ ZoneTokenAccess = NoopZoneTokenAccess{}

func (n NoopZoneTokenAccess) ValidateGenerateZoneToken(ctx context.Context, zone string, user user.User) error {
	return nil
}
