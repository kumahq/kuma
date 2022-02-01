package access

import (
	config_access "github.com/kumahq/kuma/pkg/config/access"
	"github.com/kumahq/kuma/pkg/core/access"
	"github.com/kumahq/kuma/pkg/core/user"
)

type staticZoneTokenAccess struct {
	usernames map[string]bool
	groups    map[string]bool
}

var _ ZoneTokenAccess = &staticZoneTokenAccess{}

func NewStaticZoneTokenAccess(cfg config_access.GenerateZoneTokenStaticAccessConfig) ZoneTokenAccess {
	s := &staticZoneTokenAccess{
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

func (s *staticZoneTokenAccess) ValidateGenerateZoneToken(zone string, user user.User) error {
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
