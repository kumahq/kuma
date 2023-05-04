package access

import "github.com/kumahq/kuma/pkg/core/user"

type NoopEnvoyAdminAccess struct{}

var _ EnvoyAdminAccess = NoopEnvoyAdminAccess{}

func (n NoopEnvoyAdminAccess) ValidateViewConfigDump(_ user.User) error {
	return nil
}

func (n NoopEnvoyAdminAccess) ValidateViewStats(user user.User) error {
	return nil
}

func (n NoopEnvoyAdminAccess) ValidateViewClusters(user user.User) error {
	return nil
}
