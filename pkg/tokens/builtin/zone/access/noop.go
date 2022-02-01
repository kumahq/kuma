package access

import "github.com/kumahq/kuma/pkg/core/user"

type NoopZoneTokenAccess struct {
}

var _ ZoneTokenAccess = NoopZoneTokenAccess{}

func (n NoopZoneTokenAccess) ValidateGenerateZoneToken(zone string, user user.User) error {
	return nil
}
