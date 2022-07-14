package access

import "github.com/kumahq/kuma/pkg/core/user"

type EnvoyAdminAccess interface {
	ValidateViewConfigDump(user user.User) error
	ValidateViewStats(user user.User) error
	ValidateViewClusters(user user.User) error
}
