package access

import (
	"context"

	config_access "github.com/kumahq/kuma/pkg/config/access"
	"github.com/kumahq/kuma/pkg/core/access"
	"github.com/kumahq/kuma/pkg/core/user"
)

type staticEnvoyAdminAccess struct {
	configDump accessMaps
	stats      accessMaps
	clusters   accessMaps
}

type accessMaps struct {
	usernames map[string]struct{}
	groups    map[string]struct{}
}

func (am accessMaps) Validate(user user.User) error {
	return access.Validate(am.usernames, am.groups, user, "envoy proxy info")
}

var _ EnvoyAdminAccess = &staticEnvoyAdminAccess{}

func NewStaticEnvoyAdminAccess(
	configDumpCfg config_access.ViewConfigDumpStaticAccessConfig,
	statsCfg config_access.ViewStatsStaticAccessConfig,
	clustersCfg config_access.ViewClustersStaticAccessConfig,
) EnvoyAdminAccess {
	return &staticEnvoyAdminAccess{
		configDump: mapAccess(configDumpCfg.Users, configDumpCfg.Groups),
		stats:      mapAccess(statsCfg.Users, statsCfg.Groups),
		clusters:   mapAccess(clustersCfg.Users, clustersCfg.Groups),
	}
}

func mapAccess(users, groups []string) accessMaps {
	m := accessMaps{
		usernames: make(map[string]struct{}, len(users)),
		groups:    make(map[string]struct{}, len(groups)),
	}
	for _, usr := range users {
		m.usernames[usr] = struct{}{}
	}
	for _, group := range groups {
		m.groups[group] = struct{}{}
	}
	return m
}

func (s *staticEnvoyAdminAccess) ValidateViewConfigDump(ctx context.Context, user user.User) error {
	return s.configDump.Validate(user)
}

func (s *staticEnvoyAdminAccess) ValidateViewStats(ctx context.Context, user user.User) error {
	return s.stats.Validate(user)
}

func (s *staticEnvoyAdminAccess) ValidateViewClusters(ctx context.Context, user user.User) error {
	return s.clusters.Validate(user)
}
