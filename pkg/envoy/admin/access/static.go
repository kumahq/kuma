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
	usernames map[string]bool
	groups    map[string]bool
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

func mapAccess(users []string, groups []string) accessMaps {
	m := accessMaps{
		usernames: map[string]bool{},
		groups:    map[string]bool{},
	}
	for _, usr := range users {
		m.usernames[usr] = true
	}
	for _, group := range groups {
		m.groups[group] = true
	}
	return m
}

func (s *staticEnvoyAdminAccess) ValidateViewConfigDump(ctx context.Context, user user.User) error {
	return validateAccess(s.configDump, user)
}

func (s *staticEnvoyAdminAccess) ValidateViewStats(ctx context.Context, user user.User) error {
	return validateAccess(s.stats, user)
}

func (s *staticEnvoyAdminAccess) ValidateViewClusters(ctx context.Context, user user.User) error {
	return validateAccess(s.clusters, user)
}

func validateAccess(maps accessMaps, user user.User) error {
	allowed := maps.usernames[user.Name]
	for _, group := range user.Groups {
		if maps.groups[group] {
			allowed = true
		}
	}
	if !allowed {
		return &access.AccessDeniedError{Reason: "action not allowed"}
	}
	return nil
}
