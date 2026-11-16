package access

import (
	"context"

	"github.com/kumahq/kuma/v2/pkg/core/user"
	"github.com/kumahq/kuma/v2/pkg/tokens/builtin/issuer"
)

type NoopDpTokenAccess struct{}

var _ DataplaneTokenAccess = NoopDpTokenAccess{}

func (n NoopDpTokenAccess) ValidateGenerateDataplaneToken(ctx context.Context, identity issuer.DataplaneIdentity, user user.User) error {
	return nil
}

func (n NoopDpTokenAccess) ValidateGenerateZoneIngressToken(ctx context.Context, zone string, user user.User) error {
	return nil
}
