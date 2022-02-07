package access

import "github.com/kumahq/kuma/pkg/core/user"

type ConfigDumpAccess interface {
	ValidateViewConfigDump(user user.User) error
}
