package access

import "github.com/kumahq/kuma/pkg/core/user"

type ConfigDumpAccess interface {
	ValidateGetConfigDump(user user.User) error
}
