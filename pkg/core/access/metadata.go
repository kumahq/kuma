package access

import (
	"context"

	config_access "github.com/kumahq/kuma/pkg/config/access"
	"github.com/kumahq/kuma/pkg/core/user"
)

type ControlPlaneMetadataAccess interface {
	ValidateView(ctx context.Context, user user.User) error
}

type staticMetadataAccess struct {
	usernames map[string]struct{}
	groups    map[string]struct{}
}

func NewStaticControlPlaneMetadataAccess(cfg config_access.ControlPlaneMetadataStaticAccessConfig) ControlPlaneMetadataAccess {
	s := &staticMetadataAccess{
		usernames: make(map[string]struct{}, len(cfg.Users)),
		groups:    make(map[string]struct{}, len(cfg.Groups)),
	}
	for _, u := range cfg.Users {
		s.usernames[u] = struct{}{}
	}
	for _, g := range cfg.Groups {
		s.groups[g] = struct{}{}
	}
	return s
}

func (s staticMetadataAccess) ValidateView(ctx context.Context, user user.User) error {
	return Validate(s.usernames, s.groups, user, "control-plane metadata")
}
