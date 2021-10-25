package rbac

import (
	"github.com/kumahq/kuma/pkg/core/user"
)

type GenerateDataplaneTokenAccess interface {
	ValidateGenerate(name string, mesh string, tags map[string][]string, tokenType string, user *user.User) error
}
