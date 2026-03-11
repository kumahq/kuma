package xds

import (
	"slices"

	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	util_maps "github.com/kumahq/kuma/v2/pkg/util/maps"
)

const OtelDynconfPath = "/otel"

type OtelClientLayout string

const (
	OtelClientLayoutShared    OtelClientLayout = "shared"
	OtelClientLayoutPerSignal OtelClientLayout = "per-signal"
)

type OtelSignalSource string

const (
	OtelSignalSourceExplicit OtelSignalSource = "explicit"
	OtelSignalSourceEnv      OtelSignalSource = "env"
	OtelSignalSourceMixed    OtelSignalSource = "mixed"
)

const (
	OtelBlockedReasonEnvDisabledByPlatform  = "EnvDisabledByPlatform"
	OtelBlockedReasonEnvDisabledByPolicy    = "EnvDisabledByPolicy"
	OtelBlockedReasonRequiredEnvMissing     = "RequiredEnvMissing"
	OtelBlockedReasonSignalOverridesBlocked = "SignalOverridesDisallowed"
	OtelBlockedReasonMultipleBackends       = "MultipleBackendsForSignal"
)

type OtelResolvedEnvPolicy struct {
	Mode                 motb_api.EnvMode       `json:"mode,omitempty"`
	Precedence           motb_api.EnvPrecedence `json:"precedence,omitempty"`
	AllowSignalOverrides bool                   `json:"allowSignalOverrides,omitempty"`
}

type OtelSignalRuntimePlan struct {
	Enabled         bool     `json:"enabled,omitempty"`
	EnvInputPresent bool     `json:"envInputPresent,omitempty"`
	Source          string   `json:"source,omitempty"`
	OverrideKinds   []string `json:"overrideKinds,omitempty"`
	MissingFields   []string `json:"missingFields,omitempty"`
	BlockedReasons  []string `json:"blockedReasons,omitempty"`
	RefreshInterval string   `json:"refreshInterval,omitempty"`
}

// OtelPipeBackend represents one MOTB backend for the unified /otel dynconf route.
// All signals sharing this backend use the same SocketPath.
type OtelPipeBackend struct {
	Name         string                 `json:"name,omitempty"`
	SocketPath   string                 `json:"socketPath"`
	Endpoint     string                 `json:"endpoint"`
	UseHTTP      bool                   `json:"useHTTP"`
	UseHTTPS     bool                   `json:"useHTTPS,omitempty"`
	Path         string                 `json:"path,omitempty"`
	EnvPolicy    OtelResolvedEnvPolicy  `json:"envPolicy,omitempty"`
	ClientLayout OtelClientLayout       `json:"clientLayout,omitempty"`
	Traces       *OtelSignalRuntimePlan `json:"traces,omitempty"`
	Logs         *OtelSignalRuntimePlan `json:"logs,omitempty"`
	Metrics      *OtelSignalRuntimePlan `json:"metrics,omitempty"`
}

// OtelDpConfig is sent from CP to DP via the /otel dynconf route.
type OtelDpConfig struct {
	Backends []OtelPipeBackend `json:"backends"`
}

// OtelPipeBackends accumulates backends from policy plugins during xDS generation.
// Deduplicates by backend name - all signals for the same MOTB share one socket.
type OtelPipeBackends struct {
	backends       map[string]*OtelPipeBackend        // key: backendName
	signalBackends map[OtelSignal]map[string]struct{} // key: signal -> backend names
}

func (a *OtelPipeBackends) Add(name string, b OtelPipeBackend) {
	a.mergeBase(name, b)
}

func (a *OtelPipeBackends) AddSignal(
	name string,
	b OtelPipeBackend,
	signal OtelSignal,
	plan OtelSignalRuntimePlan,
) {
	backend := a.mergeBase(name, b)
	backend.setSignalPlan(signal, cloneSignalPlan(plan))

	if a.signalBackends == nil {
		a.signalBackends = map[OtelSignal]map[string]struct{}{}
	}
	if a.signalBackends[signal] == nil {
		a.signalBackends[signal] = map[string]struct{}{}
	}
	a.signalBackends[signal][name] = struct{}{}
}

func (a *OtelPipeBackends) All() []OtelPipeBackend {
	if len(a.backends) == 0 {
		return nil
	}
	var result []OtelPipeBackend
	for _, name := range util_maps.SortedKeys(a.backends) {
		backend := cloneBackend(*a.backends[name])
		a.finalizeBackend(&backend)
		result = append(result, backend)
	}
	return result
}

func (a *OtelPipeBackends) Empty() bool {
	return len(a.backends) == 0
}

func (a *OtelPipeBackends) mergeBase(name string, b OtelPipeBackend) *OtelPipeBackend {
	if a.backends == nil {
		a.backends = map[string]*OtelPipeBackend{}
	}

	backend, ok := a.backends[name]
	if !ok {
		base := b
		base.Name = name
		backend = &base
		a.backends[name] = backend
		return backend
	}

	backend.Name = name
	backend.SocketPath = b.SocketPath
	backend.Endpoint = b.Endpoint
	backend.UseHTTP = b.UseHTTP
	backend.UseHTTPS = b.UseHTTPS
	backend.Path = b.Path
	backend.EnvPolicy = b.EnvPolicy
	return backend
}

func (a *OtelPipeBackends) finalizeBackend(backend *OtelPipeBackend) {
	if backend == nil {
		return
	}

	for _, signal := range []OtelSignal{OtelSignalTraces, OtelSignalLogs, OtelSignalMetrics} {
		plan := backend.getSignalPlan(signal)
		if plan == nil {
			continue
		}
		if len(a.signalBackends[signal]) > 1 &&
			plan.EnvInputPresent &&
			backend.EnvPolicy.Mode != motb_api.EnvModeDisabled &&
			!slices.Contains(plan.BlockedReasons, OtelBlockedReasonEnvDisabledByPlatform) &&
			!slices.Contains(plan.BlockedReasons, OtelBlockedReasonEnvDisabledByPolicy) {
			plan.BlockedReasons = appendUnique(plan.BlockedReasons, OtelBlockedReasonMultipleBackends)
			if plan.Source != "" {
				plan.Source = string(OtelSignalSourceExplicit)
			}
		}
	}

	backend.ClientLayout = OtelClientLayoutShared
	if backend.enabledSignalCount() <= 1 {
		return
	}

	for _, signal := range []OtelSignal{OtelSignalTraces, OtelSignalLogs, OtelSignalMetrics} {
		plan := backend.getSignalPlan(signal)
		if plan == nil || len(plan.OverrideKinds) == 0 {
			continue
		}
		if slices.Contains(plan.BlockedReasons, OtelBlockedReasonSignalOverridesBlocked) {
			continue
		}
		if !backend.EnvPolicy.AllowSignalOverrides {
			continue
		}
		backend.ClientLayout = OtelClientLayoutPerSignal
		return
	}
}

func (b *OtelPipeBackend) getSignalPlan(signal OtelSignal) *OtelSignalRuntimePlan {
	switch signal {
	case OtelSignalTraces:
		return b.Traces
	case OtelSignalLogs:
		return b.Logs
	case OtelSignalMetrics:
		return b.Metrics
	default:
		return nil
	}
}

func (b *OtelPipeBackend) setSignalPlan(signal OtelSignal, plan OtelSignalRuntimePlan) {
	switch signal {
	case OtelSignalTraces:
		b.Traces = &plan
	case OtelSignalLogs:
		b.Logs = &plan
	case OtelSignalMetrics:
		b.Metrics = &plan
	}
}

func (b *OtelPipeBackend) enabledSignalCount() int {
	count := 0
	for _, signal := range []OtelSignal{OtelSignalTraces, OtelSignalLogs, OtelSignalMetrics} {
		plan := b.getSignalPlan(signal)
		if plan != nil && plan.Enabled {
			count++
		}
	}
	return count
}

func cloneBackend(in OtelPipeBackend) OtelPipeBackend {
	out := in
	if in.Traces != nil {
		traces := cloneSignalPlan(*in.Traces)
		out.Traces = &traces
	}
	if in.Logs != nil {
		logs := cloneSignalPlan(*in.Logs)
		out.Logs = &logs
	}
	if in.Metrics != nil {
		metrics := cloneSignalPlan(*in.Metrics)
		out.Metrics = &metrics
	}
	return out
}

func cloneSignalPlan(in OtelSignalRuntimePlan) OtelSignalRuntimePlan {
	return OtelSignalRuntimePlan{
		Enabled:         in.Enabled,
		EnvInputPresent: in.EnvInputPresent,
		Source:          in.Source,
		OverrideKinds:   slices.Clone(in.OverrideKinds),
		MissingFields:   slices.Clone(in.MissingFields),
		BlockedReasons:  slices.Clone(in.BlockedReasons),
		RefreshInterval: in.RefreshInterval,
	}
}

func appendUnique(values []string, value string) []string {
	if slices.Contains(values, value) {
		return values
	}
	return append(values, value)
}
