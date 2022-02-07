package access

import (
	config_access "github.com/kumahq/kuma/pkg/config/access"
	"github.com/kumahq/kuma/pkg/core/access"
	"github.com/kumahq/kuma/pkg/core/user"
)

type staticConfigDumpAccess struct {
	usernames map[string]bool
	groups    map[string]bool
}

var _ ConfigDumpAccess = &staticConfigDumpAccess{}

func NewStaticConfigDumpAccess(cfg config_access.ViewConfigDumpStaticAccessConfig) ConfigDumpAccess {
	s := &staticConfigDumpAccess{
		usernames: map[string]bool{},
		groups:    map[string]bool{},
	}
	for _, usr := range cfg.Users {
		s.usernames[usr] = true
	}
	for _, group := range cfg.Groups {
		s.groups[group] = true
	}
	return s
}

func (s *staticConfigDumpAccess) ValidateViewConfigDump(user user.User) error {
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
