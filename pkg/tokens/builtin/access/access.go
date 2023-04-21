package access

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/user"
)

type DataplaneTokenAccess interface {
	ValidateGenerateDataplaneToken(ctx context.Context, name string, mesh string, tags map[string][]string, user user.User) error
	ValidateGenerateZoneIngressToken(ctx context.Context, zone string, user user.User) error
}
