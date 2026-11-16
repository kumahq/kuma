package access

import (
	"context"

	"github.com/kumahq/kuma/v3/pkg/core/user"
	"github.com/kumahq/kuma/v3/pkg/tokens/builtin/issuer"
)

type DataplaneTokenAccess interface {
	ValidateGenerateDataplaneToken(ctx context.Context, identity issuer.DataplaneIdentity, user user.User) error
}
