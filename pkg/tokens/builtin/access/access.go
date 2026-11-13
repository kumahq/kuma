package access

import (
	"context"

	"github.com/kumahq/kuma/v2/pkg/core/user"
	"github.com/kumahq/kuma/v2/pkg/tokens/builtin/issuer"
)

type DataplaneTokenAccess interface {
	ValidateGenerateDataplaneToken(ctx context.Context, identity issuer.DataplaneIdentity, user user.User) error
	ValidateGenerateZoneIngressToken(ctx context.Context, zone string, user user.User) error
}
