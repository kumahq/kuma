package access

import (
	"context"
	"fmt"

	config_access "github.com/kumahq/kuma/v2/pkg/config/access"
	"github.com/kumahq/kuma/v2/pkg/core/access"
	"github.com/kumahq/kuma/v2/pkg/core/user"
	"github.com/kumahq/kuma/v2/pkg/tokens/builtin/zone"
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

func (s *staticZoneTokenAccess) ValidateGenerateZoneToken(ctx context.Context, identity zone.Identity, user user.User) error {
	return access.Validate(s.usernames, s.groups, user, fmt.Sprintf("generate zone token for zone '%s'", identity.Zone))
}
