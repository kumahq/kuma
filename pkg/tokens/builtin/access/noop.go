package access

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/user"
)

type NoopDpTokenAccess struct{}

var _ DataplaneTokenAccess = NoopDpTokenAccess{}

func (n NoopDpTokenAccess) ValidateGenerateDataplaneToken(ctx context.Context, name, mesh string, tags map[string][]string, user user.User) error {
	return nil
}

func (n NoopDpTokenAccess) ValidateGenerateZoneIngressToken(ctx context.Context, zone string, user user.User) error {
	return nil
}
