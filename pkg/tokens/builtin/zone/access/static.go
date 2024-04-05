package access

import (
	"context"
	"fmt"

	config_access "github.com/kumahq/kuma/pkg/config/access"
	"github.com/kumahq/kuma/pkg/core/access"
	"github.com/kumahq/kuma/pkg/core/user"
)

type staticZoneTokenAccess struct {
	usernames map[string]struct{}
	groups    map[string]struct{}
}

var _ ZoneTokenAccess = &staticZoneTokenAccess{}

func NewStaticZoneTokenAccess(cfg config_access.GenerateZoneTokenStaticAccessConfig) ZoneTokenAccess {
	s := &staticZoneTokenAccess{
		usernames: map[string]struct{}{},
		groups:    map[string]struct{}{},
	}
	for _, user := range cfg.Users {
		s.usernames[user] = struct{}{}
	}
	for _, group := range cfg.Groups {
		s.groups[group] = struct{}{}
	}
	return s
}

func (s *staticZoneTokenAccess) ValidateGenerateZoneToken(ctx context.Context, zone string, user user.User) error {
	return access.Validate(s.usernames, s.groups, user, fmt.Sprintf("generate zone token for zone '%s'", zone))
}
