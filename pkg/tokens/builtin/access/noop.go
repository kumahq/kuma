package access

import (
	"context"

	"github.com/kumahq/kuma/v2/pkg/core/user"
)

type NoopDpTokenAccess struct{}

var _ DataplaneTokenAccess = NoopDpTokenAccess{}

func (NoopDpTokenAccess) ValidateGenerateDataplaneToken(ctx context.Context, name string, mesh string, tags map[string][]string, user user.User) error {
	return nil
}

func (NoopDpTokenAccess) ValidateGenerateZoneIngressToken(ctx context.Context, zone string, user user.User) error {
	return nil
}
