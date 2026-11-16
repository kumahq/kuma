package access

import (
	"context"

	config_access "github.com/kumahq/kuma/v3/pkg/config/access"
	"github.com/kumahq/kuma/v3/pkg/core/access"
	"github.com/kumahq/kuma/v3/pkg/core/user"
	"github.com/kumahq/kuma/v3/pkg/tokens/builtin/issuer"
)

type staticGenerateDataplaneTokenAccess struct {
	usernames map[string]struct{}
	groups    map[string]struct{}
}

var _ DataplaneTokenAccess = &staticGenerateDataplaneTokenAccess{}

func NewStaticGenerateDataplaneTokenAccess(cfg config_access.GenerateDPTokenStaticAccessConfig) DataplaneTokenAccess {
	s := &staticGenerateDataplaneTokenAccess{
		usernames: make(map[string]struct{}, len(cfg.Users)),
		groups:    make(map[string]struct{}, len(cfg.Groups)),
	}
	for _, u := range cfg.Users {
		s.usernames[u] = struct{}{}
	}
	for _, group := range cfg.Groups {
		s.groups[group] = struct{}{}
	}
	return s
}

func (s *staticGenerateDataplaneTokenAccess) ValidateGenerateDataplaneToken(ctx context.Context, identity issuer.DataplaneIdentity, user user.User) error {
	return access.Validate(s.usernames, s.groups, user, "generate dataplane token")
}
