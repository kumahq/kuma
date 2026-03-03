package xds

import (
	util_maps "github.com/kumahq/kuma/v2/pkg/util/maps"
)

const OtelDynconfPath = "/otel"

// OtelPipeBackend represents one MOTB backend for the unified /otel dynconf route.
// All signals sharing this backend use the same SocketPath.
type OtelPipeBackend struct {
	SocketPath string `json:"socketPath"`
	Endpoint   string `json:"endpoint"`
	UseHTTP    bool   `json:"useHTTP"`
	Path       string `json:"path,omitempty"`
}

// OtelDpConfig is sent from CP to DP via the /otel dynconf route.
type OtelDpConfig struct {
	Backends []OtelPipeBackend `json:"backends"`
}

// OtelPipeBackends accumulates backends from policy plugins during xDS generation.
// Deduplicates by backend name - all signals for the same MOTB share one socket.
type OtelPipeBackends struct {
	backends map[string]OtelPipeBackend // key: backendName
}

func (a *OtelPipeBackends) Add(name string, b OtelPipeBackend) {
	if a.backends == nil {
		a.backends = map[string]OtelPipeBackend{}
	}
	a.backends[name] = b
}

func (a *OtelPipeBackends) All() []OtelPipeBackend {
	if len(a.backends) == 0 {
		return nil
	}
	var result []OtelPipeBackend
	for _, name := range util_maps.SortedKeys(a.backends) {
		result = append(result, a.backends[name])
	}
	return result
}

func (a *OtelPipeBackends) Empty() bool {
	return len(a.backends) == 0
}
