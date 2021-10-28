package rbac

import (
	config_rbac "github.com/kumahq/kuma/pkg/config/rbac"
	"github.com/kumahq/kuma/pkg/core/rbac"
	"github.com/kumahq/kuma/pkg/core/user"
)

type staticGenerateUserTokenAccess struct {
	usernames map[string]bool
	groups    map[string]bool
}

var _ GenerateUserTokenAccess = &staticGenerateUserTokenAccess{}

func NewStaticGenerateUserTokenAccess(cfg config_rbac.GenerateUserTokenRBACStaticConfig) GenerateUserTokenAccess {
	s := &staticGenerateUserTokenAccess{
		usernames: map[string]bool{},
		groups:    map[string]bool{},
	}
	for _, user := range cfg.Users {
		s.usernames[user] = true
	}
	for _, group := range cfg.Groups {
		s.groups[group] = true
	}
	return s
}

func (s *staticGenerateUserTokenAccess) ValidateGenerate(user user.User) error {
	allowed := s.usernames[user.Name]
	for _, group := range user.Groups {
		if s.groups[group] {
			allowed = true
		}
	}
	if !allowed {
		return &rbac.AccessDeniedError{Reason: "action not allowed"}
	}
	return nil
}
