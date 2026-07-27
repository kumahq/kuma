package v1alpha1

import (
	"testing"
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
