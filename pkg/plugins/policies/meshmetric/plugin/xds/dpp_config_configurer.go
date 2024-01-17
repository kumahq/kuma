package xds

import (
	"encoding/json"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

type DppConfigConfigurer struct {
	ListenerName string
	DpConfig     MeshMetricDpConfig
}

func (dcc *DppConfigConfigurer) ConfigureListener(proxy *core_xds.Proxy) (envoy_common.NamedResource, error) {
	marshal, err := json.Marshal(dcc.DpConfig)
	if err != nil {
		return nil, err
	}

	return envoy_listeners.NewListenerBuilder(proxy.APIVersion, dcc.ListenerName).
		Configure(envoy_listeners.PipeListener(core_xds.MeshMetricsDynamicConfigurationSocketName(proxy.Metadata.WorkDir))).
		Configure(envoy_listeners.FilterChain(
			envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
				Configure(
					envoy_listeners.DirectResponse(dcc.ListenerName, []v3.DirectResponseEndpoints{{
						Path:       "/",
						StatusCode: 200,
						Response:   string(marshal),
					}}),
				),
		)).
		Build()
}

type MeshMetricDpConfig struct {
	Observability Observability `json:"observability"`
}

type Observability struct {
	Metrics Metrics `json:"metrics"`
}

type Metrics struct {
	Applications []Application `json:"applications"`
	Backends     []Backend     `json:"backends"`
}

type Application struct {
	Path    string `json:"path"`
	Port    uint32 `json:"port"`
	Address string `json:"address"`
}

type Backend struct {
	Type          string                `json:"type"`
	OpenTelemetry *OpenTelemetryBackend `json:"openTelemetry,omitempty"`
}

type OpenTelemetryBackend struct {
	Endpoint string `json:"endpoint"`
}
