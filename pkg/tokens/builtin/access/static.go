package access

import (
	"context"

	config_access "github.com/kumahq/kuma/pkg/config/access"
	"github.com/kumahq/kuma/pkg/core/access"
	"github.com/kumahq/kuma/pkg/core/user"
)

type staticGenerateDataplaneTokenAccess struct {
	usernames map[string]bool
	groups    map[string]bool
}

var _ DataplaneTokenAccess = &staticGenerateDataplaneTokenAccess{}

func NewStaticGenerateDataplaneTokenAccess(cfg config_access.GenerateDPTokenStaticAccessConfig) DataplaneTokenAccess {
	s := &staticGenerateDataplaneTokenAccess{
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

func (s *staticGenerateDataplaneTokenAccess) ValidateGenerateDataplaneToken(ctx context.Context, name string, mesh string, tags map[string][]string, user user.User) error {
	return s.validateAccess(user)
}

func (s *staticGenerateDataplaneTokenAccess) ValidateGenerateZoneIngressToken(ctx context.Context, zone string, user user.User) error {
	return s.validateAccess(user)
}

func (s *staticGenerateDataplaneTokenAccess) validateAccess(user user.User) error {
	allowed := s.usernames[user.Name]
	for _, group := range user.Groups {
		if s.groups[group] {
			allowed = true
		}
	}
	if !allowed {
		return &access.AccessDeniedError{Reason: "action not allowed"}
	}
	return nil
}
