package access

import "github.com/kumahq/kuma/pkg/core/user"

type NoopConfigDumpAccess struct {
}

var _ ConfigDumpAccess = NoopConfigDumpAccess{}

func (n NoopConfigDumpAccess) ValidateViewConfigDump(_ user.User) error {
	return nil
}
