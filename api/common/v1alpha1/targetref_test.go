package v1alpha1

import (
	"testing"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
)

func TestIncludesGateways(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		ref      TargetRef
		expected bool
	}{
		"mesh includes gateways": {
			ref:      TargetRef{Kind: Mesh},
			expected: true,
		},
		"mesh http route includes gateways": {
			ref:      TargetRef{Kind: MeshHTTPRoute},
			expected: true,
		},
		"dataplane without labels excludes gateways": {
			ref:      TargetRef{Kind: Dataplane},
			expected: false,
		},
		"dataplane with non proxy type labels excludes gateways": {
			ref: TargetRef{
				Kind:   Dataplane,
				Labels: &map[string]string{"kuma.io/service": "backend"},
			},
			expected: false,
		},
		"dataplane with a gateway proxy-type label still excludes gateways": {
			// Kind: Dataplane never distinguished gateways even when
			// proxyTypes existed (the field was never valid on it), and
			// this label doesn't resurrect that: a targetRef can no longer
			// scope a policy's rules[]/to[] to gateways-only at all.
			ref: TargetRef{
				Kind:   Dataplane,
				Labels: &map[string]string{mesh_proto.ProxyTypeLabel: string(mesh_proto.GatewayLabel)},
			},
			expected: false,
		},
		"mesh service excludes gateways": {
			ref:      TargetRef{Kind: MeshService},
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := IncludesGateways(tt.ref); got != tt.expected {
				t.Fatalf("IncludesGateways() = %v, want %v", got, tt.expected)
			}
		})
	}
}
