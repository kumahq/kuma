package access

import (
	"context"

	"github.com/kumahq/kuma/v3/pkg/core/user"
)

type EnvoyAdminAccess interface {
	ValidateViewConfigDump(ctx context.Context, user user.User) error
	ValidateViewStats(ctx context.Context, user user.User) error
	ValidateViewClusters(ctx context.Context, user user.User) error
}
