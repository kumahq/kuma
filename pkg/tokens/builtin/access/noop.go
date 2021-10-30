package access

import "github.com/kumahq/kuma/pkg/core/user"

type NoopGenerateDpTokenAccess struct {
}

var _ GenerateDataplaneTokenAccess = NoopGenerateDpTokenAccess{}

func (n NoopGenerateDpTokenAccess) ValidateGenerate(name string, mesh string, tags map[string][]string, tokenType string, user user.User) error {
	return nil
}
