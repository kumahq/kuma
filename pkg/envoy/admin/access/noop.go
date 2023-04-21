package access

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/user"
)

type NoopEnvoyAdminAccess struct{}

var _ EnvoyAdminAccess = NoopEnvoyAdminAccess{}

func (n NoopEnvoyAdminAccess) ValidateViewConfigDump(_ context.Context, _ user.User) error {
	return nil
}

func (n NoopEnvoyAdminAccess) ValidateViewStats(_ context.Context, user user.User) error {
	return nil
}

func (n NoopEnvoyAdminAccess) ValidateViewClusters(_ context.Context, user user.User) error {
	return nil
}
