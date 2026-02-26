package xds

import (
	"testing"

	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

func TestCollectorEndpointString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint *core_xds.Endpoint
		expected string
	}{
		{
			name: "ipv4 host and port",
			endpoint: &core_xds.Endpoint{
				Target: "10.0.0.1",
				Port:   4317,
			},
			expected: "10.0.0.1:4317",
		},
		{
			name: "ipv6 host and port",
			endpoint: &core_xds.Endpoint{
				Target: "2001:db8::1",
				Port:   4318,
			},
			expected: "[2001:db8::1]:4318",
		},
		{
			name: "host without port",
			endpoint: &core_xds.Endpoint{
				Target: "collector.mesh",
			},
			expected: "collector.mesh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := collectorEndpointString(tt.endpoint); got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
