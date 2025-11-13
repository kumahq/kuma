package access

import "github.com/kumahq/kuma/v2/pkg/core/user"

type GenerateUserTokenAccess interface {
	ValidateGenerate(user user.User) error
}
