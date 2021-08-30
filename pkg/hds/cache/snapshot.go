package cache

import (
	envoy_service_health_v3 "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/pkg/errors"

	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

const HealthCheckSpecifierType = "envoy.service.health.v3.HealthCheckSpecifier"

func NewSnapshot(version string, hcs *envoy_service_health_v3.HealthCheckSpecifier) util_xds_v3.Snapshot {
	return &Snapshot{
		HealthChecks: cache.Resources{
			Version: version,
			Items: map[string]envoy_types.ResourceWithTTL{
				"hcs": {Resource: hcs},
			},
		},
	}
}

// Snapshot is an internally consistent snapshot of HDS resources.
type Snapshot struct {
	HealthChecks cache.Resources
}

func (s *Snapshot) GetSupportedTypes() []string {
	return []string{HealthCheckSpecifierType}
}

func (s *Snapshot) Consistent() error {
	if s == nil {
		return errors.New("nil Snapshot")
	}
	return nil
}

func (s *Snapshot) GetResources(typ string) map[string]envoy_types.Resource {
	if s == nil || typ != HealthCheckSpecifierType {
		return nil
	}
	withoutTtl := make(map[string]envoy_types.Resource, len(s.HealthChecks.Items))
	for name, res := range s.HealthChecks.Items {
		withoutTtl[name] = res.Resource
	}
	return withoutTtl
}

func (s *Snapshot) GetVersion(typ string) string {
	if s == nil || typ != HealthCheckSpecifierType {
		return ""
	}
	return s.HealthChecks.Version
}

func (s *Snapshot) WithVersion(typ string, version string) util_xds_v3.Snapshot {
	if s == nil {
		return nil
	}
	if s.GetVersion(typ) == version || typ != HealthCheckSpecifierType {
		return s
	}
	n := cache.Resources{
		Version: version,
		Items:   s.HealthChecks.Items,
	}
	return &Snapshot{HealthChecks: n}
}
