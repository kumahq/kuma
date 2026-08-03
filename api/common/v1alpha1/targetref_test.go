package v1alpha1

import (
	"encoding/json"
	"reflect"
	"testing"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
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

func TestBackendRefReferencesRealObject(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		ref      BackendRef
		expected bool
	}{
		"mesh service by name is real": {
			ref: BackendRef{
				TargetRef: TargetRef{
					Kind:   MeshService,
					Labels: pointer.To(map[string]string{mesh_proto.DisplayName: "backend"}),
				},
			},
			expected: true,
		},
		"mesh external service is real": {
			ref: BackendRef{
				TargetRef: TargetRef{
					Kind:   MeshExternalService,
					Labels: pointer.To(map[string]string{mesh_proto.DisplayName: "payments"}),
				},
			},
			expected: true,
		},
		"mesh multizone service is real": {
			ref: BackendRef{
				TargetRef: TargetRef{
					Kind:   MeshMultiZoneService,
					Labels: pointer.To(map[string]string{mesh_proto.DisplayName: "global-backend"}),
				},
			},
			expected: true,
		},
		"legacy mesh service subset is not real": {
			ref: BackendRef{
				TargetRef: TargetRef{
					Kind:   LegacyMeshServiceSubsetKind(),
					Labels: pointer.To(map[string]string{mesh_proto.DisplayName: "backend"}),
				},
			},
			expected: false,
		},
		"empty ref is not real": {
			ref:      BackendRef{},
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tt.ref.ReferencesRealObject(); got != tt.expected {
				t.Fatalf("ReferencesRealObject() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBackendRefRealResourceSelectorDefaultsNamespaceForLabels(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		ref            BackendRef
		defaultNS      string
		expectedLabels map[string]string
	}{
		"MeshService": {
			ref: BackendRef{
				TargetRef: TargetRef{
					Kind: MeshService,
					Labels: pointer.To(map[string]string{
						mesh_proto.DisplayName:      "backend",
						mesh_proto.KubeNamespaceTag: "team-a",
					}),
				},
			},
			defaultNS: "ignored",
			expectedLabels: map[string]string{
				mesh_proto.DisplayName:      "backend",
				mesh_proto.KubeNamespaceTag: "team-a",
			},
		},
		"MeshExternalService": {
			ref: BackendRef{
				TargetRef: TargetRef{
					Kind: MeshExternalService,
					Labels: pointer.To(map[string]string{
						mesh_proto.DisplayName:      "payments",
						mesh_proto.KubeNamespaceTag: "team-a",
					}),
				},
			},
			defaultNS: "ignored",
			expectedLabels: map[string]string{
				mesh_proto.DisplayName:      "payments",
				mesh_proto.KubeNamespaceTag: "team-a",
			},
		},
		"MeshMultiZoneService": {
			ref: BackendRef{
				TargetRef: TargetRef{
					Kind: MeshMultiZoneService,
					Labels: pointer.To(map[string]string{
						mesh_proto.DisplayName:      "global-backend",
						mesh_proto.KubeNamespaceTag: "team-a",
					}),
				},
			},
			defaultNS: "ignored",
			expectedLabels: map[string]string{
				mesh_proto.DisplayName:      "global-backend",
				mesh_proto.KubeNamespaceTag: "team-a",
			},
		},
		"MeshService injects default namespace when only display-name is set": {
			ref: BackendRef{
				TargetRef: TargetRef{
					Kind: MeshService,
					Labels: pointer.To(map[string]string{
						mesh_proto.DisplayName: "backend",
					}),
				},
			},
			defaultNS: "team-a",
			expectedLabels: map[string]string{
				mesh_proto.DisplayName:      "backend",
				mesh_proto.KubeNamespaceTag: "team-a",
			},
		},
		"MeshService labels are cloned": {
			ref: BackendRef{
				TargetRef: TargetRef{
					Kind: MeshService,
					Labels: pointer.To(map[string]string{
						mesh_proto.DisplayName:      "foo_bar_svc",
						mesh_proto.KubeNamespaceTag: "explicit-ns",
					}),
				},
			},
			defaultNS: "ignored",
			expectedLabels: map[string]string{
				mesh_proto.DisplayName:      "foo_bar_svc",
				mesh_proto.KubeNamespaceTag: "explicit-ns",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			labels, sectionName, ok := tt.ref.RealResourceSelector(tt.defaultNS)
			if !ok {
				t.Fatal("RealResourceSelector() returned ok=false")
			}
			if sectionName != "" {
				t.Fatalf("RealResourceSelector() sectionName = %q, want empty", sectionName)
			}
			if !reflect.DeepEqual(labels, tt.expectedLabels) {
				t.Fatalf("RealResourceSelector() labels = %v, want %v", labels, tt.expectedLabels)
			}
		})
	}
}

func TestBackendRefRealResourceSelectorKeepsExplicitSectionNameOverPort(t *testing.T) {
	t.Parallel()

	ref := BackendRef{
		TargetRef: TargetRef{
			Kind:        MeshService,
			Labels:      pointer.To(map[string]string{mesh_proto.DisplayName: "backend"}),
			SectionName: pointer.To("http"),
		},
		Port: pointer.To(uint32(80)),
	}

	_, sectionName, ok := ref.RealResourceSelector("kuma-demo")
	if !ok {
		t.Fatal("RealResourceSelector() returned ok=false")
	}
	if sectionName != "http" {
		t.Fatalf("RealResourceSelector() sectionName = %q, want http", sectionName)
	}
}

func TestBackendRefRealResourceSelectorRejectsRealRefsWithoutLabels(t *testing.T) {
	t.Parallel()

	ref := BackendRef{
		TargetRef: TargetRef{
			Kind: MeshService,
		},
	}

	if _, _, ok := ref.RealResourceSelector("default"); ok {
		t.Fatal("RealResourceSelector() returned ok=true, want false")
	}
}

func TestTargetRefUnmarshalIgnoresLegacyNameNamespaceFields(t *testing.T) {
	t.Parallel()

	var ref TargetRef
	err := json.Unmarshal([]byte(`{
		"kind": "MeshService",
		"name": "backend",
		"namespace": "team-a",
		"sectionName": "http"
	}`), &ref)
	if err != nil {
		t.Fatal(err)
	}

	if ref.Labels != nil {
		t.Fatalf("labels = %v, want nil", pointer.Deref(ref.Labels))
	}
	if pointer.Deref(ref.SectionName) != "http" {
		t.Fatalf("sectionName = %q, want http", pointer.Deref(ref.SectionName))
	}
}

func TestBackendRefUnmarshalIgnoresLegacyNameAndKeepsBackendFields(t *testing.T) {
	t.Parallel()

	var ref BackendRef
	err := json.Unmarshal([]byte(`{
		"kind": "MeshService",
		"name": "backend",
		"weight": 7,
		"port": 8080
	}`), &ref)
	if err != nil {
		t.Fatal(err)
	}

	if ref.Labels != nil {
		t.Fatalf("labels = %v, want nil", pointer.Deref(ref.Labels))
	}
	if _, _, ok := ref.RealResourceSelector("kuma-demo"); ok {
		t.Fatal("RealResourceSelector() returned ok=true, want false")
	}
	if pointer.Deref(ref.Weight) != 7 {
		t.Fatalf("weight = %d, want 7", pointer.Deref(ref.Weight))
	}
	if pointer.Deref(ref.Port) != 8080 {
		t.Fatalf("port = %d, want 8080", pointer.Deref(ref.Port))
	}
}

func TestBackendRefHashUsesRealResourceLabels(t *testing.T) {
	t.Parallel()

	t.Run("real backend refs ignore label ordering when labels match", func(t *testing.T) {
		t.Parallel()

		a := BackendRef{
			TargetRef: TargetRef{
				Kind:   MeshService,
				Labels: &map[string]string{mesh_proto.DisplayName: "backend", mesh_proto.KubeNamespaceTag: "kuma-demo"},
			},
			Port: pointer.To(uint32(8080)),
		}
		b := BackendRef{
			TargetRef: TargetRef{
				Kind:   MeshService,
				Labels: &map[string]string{mesh_proto.KubeNamespaceTag: "kuma-demo", mesh_proto.DisplayName: "backend"},
			},
			Port: pointer.To(uint32(8080)),
		}

		if a.Hash() != b.Hash() {
			t.Fatalf("Hash() should be based on real-resource labels, got %q and %q", a.Hash(), b.Hash())
		}
	})

	t.Run("legacy backend refs hash service identity from labels", func(t *testing.T) {
		t.Parallel()

		base := BackendRef{
			TargetRef: TargetRef{
				Kind:   LegacyMeshServiceSubsetKind(),
				Labels: pointer.To(map[string]string{mesh_proto.DisplayName: "backend"}),
			},
			Port: pointer.To(uint32(8080)),
		}
		otherPort := BackendRef{
			TargetRef: TargetRef{
				Kind:   LegacyMeshServiceSubsetKind(),
				Labels: pointer.To(map[string]string{mesh_proto.DisplayName: "backend"}),
			},
			Port: pointer.To(uint32(9090)),
		}

		if base.Hash() == otherPort.Hash() {
			t.Fatalf("Hash() should distinguish derived ports, got %q", base.Hash())
		}
	})
}
