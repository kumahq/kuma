package status

import (
	"slices"
	"sync"

	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

const (
	SignalStateReady     = "ready"
	SignalStateBlocked   = "blocked"
	SignalStateMissing   = "missing"
	SignalStateAmbiguous = "ambiguous"
)

type Cache struct {
	mu       sync.RWMutex
	statuses map[core_model.ResourceKey]*mesh_proto.DataplaneInsight_OpenTelemetry
}

func NewCache() *Cache {
	return &Cache{
		statuses: map[core_model.ResourceKey]*mesh_proto.DataplaneInsight_OpenTelemetry{},
	}
}

func (c *Cache) Set(key core_model.ResourceKey, status *mesh_proto.DataplaneInsight_OpenTelemetry) {
	if c == nil {
		return
	}

	if status == nil {
		c.mu.Lock()
		delete(c.statuses, key)
		c.mu.Unlock()
		return
	}

	cloned := cloneStatus(status)

	c.mu.Lock()
	c.statuses[key] = cloned
	c.mu.Unlock()
}

func (c *Cache) Get(key core_model.ResourceKey) *mesh_proto.DataplaneInsight_OpenTelemetry {
	if c == nil {
		return nil
	}

	c.mu.RLock()
	status := c.statuses[key]
	c.mu.RUnlock()

	return cloneStatus(status)
}

func Build(backends []core_xds.OtelPipeBackend) *mesh_proto.DataplaneInsight_OpenTelemetry {
	if len(backends) == 0 {
		return nil
	}

	result := make([]*mesh_proto.DataplaneInsight_OpenTelemetry_Backend, 0, len(backends))
	for _, backend := range backends {
		result = append(result, &mesh_proto.DataplaneInsight_OpenTelemetry_Backend{
			Name:    backend.Name,
			Traces:  buildSignalStatus(backend, backend.Traces),
			Logs:    buildSignalStatus(backend, backend.Logs),
			Metrics: buildSignalStatus(backend, backend.Metrics),
		})
	}

	return &mesh_proto.DataplaneInsight_OpenTelemetry{
		Backends: result,
	}
}

func buildSignalStatus(
	backend core_xds.OtelPipeBackend,
	plan *core_xds.OtelSignalRuntimePlan,
) *mesh_proto.DataplaneInsight_OpenTelemetry_Signal {
	if plan == nil {
		return nil
	}

	return &mesh_proto.DataplaneInsight_OpenTelemetry_Signal{
		Enabled:         plan.Enabled,
		EnvAllowed:      envAllowed(backend, plan),
		EnvInputPresent: plan.EnvInputPresent,
		State:           signalState(plan),
		OverrideKinds:   slices.Clone(plan.OverrideKinds),
		MissingFields:   slices.Clone(plan.MissingFields),
		BlockedReasons:  slices.Clone(plan.BlockedReasons),
	}
}

func envAllowed(backend core_xds.OtelPipeBackend, plan *core_xds.OtelSignalRuntimePlan) bool {
	if backend.EnvPolicy == nil {
		return true
	}
	return backend.EnvPolicy.Mode != motb_api.EnvModeDisabled &&
		!slices.Contains(plan.BlockedReasons, core_xds.OtelBlockedReasonEnvDisabledByPolicy)
}

func signalState(plan *core_xds.OtelSignalRuntimePlan) string {
	switch {
	case slices.Contains(plan.BlockedReasons, core_xds.OtelBlockedReasonMultipleBackends):
		return SignalStateAmbiguous
	case len(plan.MissingFields) > 0,
		slices.Contains(plan.BlockedReasons, core_xds.OtelBlockedReasonRequiredEnvMissing):
		return SignalStateMissing
	case len(plan.BlockedReasons) > 0 && hasHardBlockedReason(plan):
		return SignalStateBlocked
	default:
		return SignalStateReady
	}
}

func hasHardBlockedReason(plan *core_xds.OtelSignalRuntimePlan) bool {
	for _, reason := range plan.BlockedReasons {
		switch reason {
		case core_xds.OtelBlockedReasonEnvDisabledByPolicy,
			core_xds.OtelBlockedReasonSignalOverridesBlocked:
			continue
		default:
			return true
		}
	}
	return false
}

func cloneStatus(status *mesh_proto.DataplaneInsight_OpenTelemetry) *mesh_proto.DataplaneInsight_OpenTelemetry {
	if status == nil {
		return nil
	}
	return proto.CloneOf(status)
}
