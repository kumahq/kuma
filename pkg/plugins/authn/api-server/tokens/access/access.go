package access

import "github.com/kumahq/kuma/pkg/core/user"

type GenerateUserTokenAccess interface {
	ValidateGenerate(user user.User) error
}
