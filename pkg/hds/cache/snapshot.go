package cache

import (
	envoy_service_health_v3 "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/pkg/errors"
)

const HealthCheckSpecifierType = "envoy.service.health.v3.HealthCheckSpecifier"

func NewSnapshot(hcs *envoy_service_health_v3.HealthCheckSpecifier) Snapshot {
	return &snapshot{
		HealthChecks: cache.Resources{
			Items: map[string]envoy_types.Resource{
				"hcs": hcs,
			},
		},
	}
}

// Snapshot is an internally consistent snapshot of HDS resources.
type snapshot struct {
	HealthChecks cache.Resources
}

func (s *snapshot) GetSupportedTypes() []string {
	return []string{HealthCheckSpecifierType}
}

func (s *snapshot) Consistent() error {
	if s == nil {
		return errors.New("nil snapshot")
	}
	return nil
}

func (s *snapshot) GetResources(typ string) map[string]envoy_types.Resource {
	if s == nil || typ != HealthCheckSpecifierType {
		return nil
	}
	return s.HealthChecks.Items
}

func (s *snapshot) GetVersion(typ string) string {
	if s == nil || typ != HealthCheckSpecifierType {
		return ""
	}
	return s.HealthChecks.Version
}

func (s *snapshot) WithVersion(typ string, version string) Snapshot {
	if s == nil {
		return nil
	}
	if s.GetVersion(typ) == version {
		return s
	}
	n := cache.Resources{
		Version: version,
		Items:   s.HealthChecks.Items,
	}
	return &snapshot{HealthChecks: n}
}
