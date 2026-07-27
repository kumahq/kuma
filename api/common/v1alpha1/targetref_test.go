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
		"dataplane without labels includes gateways": {
			ref:      TargetRef{Kind: Dataplane},
			expected: true,
		},
		"dataplane with non proxy type labels includes gateways": {
			ref: TargetRef{
				Kind:   Dataplane,
				Labels: &map[string]string{"kuma.io/service": "backend"},
			},
			expected: true,
		},
		"dataplane with gateway proxy type includes gateways": {
			ref: TargetRef{
				Kind:   Dataplane,
				Labels: &map[string]string{mesh_proto.ProxyTypeLabel: string(mesh_proto.GatewayLabel)},
			},
			expected: true,
		},
		"dataplane with sidecar proxy type excludes gateways": {
			ref: TargetRef{
				Kind:   Dataplane,
				Labels: &map[string]string{mesh_proto.ProxyTypeLabel: string(mesh_proto.SidecarLabel)},
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
