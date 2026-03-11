package containers

import (
	"testing"
	"time"

	kube_core "k8s.io/api/core/v1"

	runtime_k8s "github.com/kumahq/kuma/v2/pkg/config/plugins/runtime/k8s"
	config_types "github.com/kumahq/kuma/v2/pkg/config/types"
)

func TestDataplaneProxyFactory_sidecarEnvVars_injectsOtelEnvFlagWhenEnabled(t *testing.T) {
	factory := &DataplaneProxyFactory{
		ContainerConfig: runtime_k8s.DataplaneContainer{
			DrainTime: config_types.Duration{Duration: 30 * time.Second},
			EnvVars:   map[string]string{},
		},
		BuiltinDNS:      runtime_k8s.BuiltinDNS{},
		otelPipeEnabled: true,
		otelEnvEnabled:  true,
	}

	envVars, err := factory.sidecarEnvVars("default", nil)
	if err != nil {
		t.Fatalf("sidecarEnvVars() error = %v", err)
	}

	envVar, ok := findEnvVar(envVars, "KUMA_DATAPLANE_RUNTIME_OTEL_ENV_ENABLED")
	if !ok {
		t.Fatalf("expected KUMA_DATAPLANE_RUNTIME_OTEL_ENV_ENABLED to be injected")
	}
	if envVar.Value != "true" {
		t.Fatalf("expected KUMA_DATAPLANE_RUNTIME_OTEL_ENV_ENABLED=true, got %q", envVar.Value)
	}
}

func TestDataplaneProxyFactory_sidecarEnvVars_skipsOtelEnvFlagWhenDisabled(t *testing.T) {
	factory := &DataplaneProxyFactory{
		ContainerConfig: runtime_k8s.DataplaneContainer{
			DrainTime: config_types.Duration{Duration: 30 * time.Second},
			EnvVars:   map[string]string{},
		},
		BuiltinDNS:      runtime_k8s.BuiltinDNS{},
		otelPipeEnabled: true,
		otelEnvEnabled:  false,
	}

	envVars, err := factory.sidecarEnvVars("default", nil)
	if err != nil {
		t.Fatalf("sidecarEnvVars() error = %v", err)
	}

	if _, ok := findEnvVar(envVars, "KUMA_DATAPLANE_RUNTIME_OTEL_ENV_ENABLED"); ok {
		t.Fatalf("expected KUMA_DATAPLANE_RUNTIME_OTEL_ENV_ENABLED to stay unset")
	}
}

func findEnvVar(envVars []kube_core.EnvVar, name string) (kube_core.EnvVar, bool) {
	for _, envVar := range envVars {
		if envVar.Name == name {
			return envVar, true
		}
	}
	return kube_core.EnvVar{}, false
}
