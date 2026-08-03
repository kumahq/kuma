// +kubebuilder:object:generate=true
package v1alpha1

import (
	"fmt"
	"maps"
	"sort"
	"strings"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	util_maps "github.com/kumahq/kuma/v3/pkg/util/maps"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

type TargetRefKind string

var (
	Mesh                 TargetRefKind = "Mesh"
	Dataplane            TargetRefKind = "Dataplane"
	MeshService          TargetRefKind = "MeshService"
	MeshExternalService  TargetRefKind = "MeshExternalService"
	MeshMultiZoneService TargetRefKind = "MeshMultiZoneService"
	MeshHTTPRoute        TargetRefKind = "MeshHTTPRoute"
)

// meshSubset and meshServiceSubset are legacy kinds that predate real
// resources (MeshService, MeshExternalService, MeshMultiZoneService). They
// stay unexported: the wire values are still valid and must keep their
// current validation/matching behavior, but no new Go code should reference
// them directly.
const (
	meshSubset        TargetRefKind = "MeshSubset"
	meshServiceSubset TargetRefKind = "MeshServiceSubset"
)

// LegacyMeshSubsetKind returns the legacy MeshSubset wire value without
// re-exporting the kind constant.
func LegacyMeshSubsetKind() TargetRefKind {
	return meshSubset
}

// LegacyMeshServiceSubsetKind returns the legacy MeshServiceSubset wire value
// without re-exporting the kind constant.
func LegacyMeshServiceSubsetKind() TargetRefKind {
	return meshServiceSubset
}

var order = map[TargetRefKind]int{
	Mesh:                 1,
	Dataplane:            2,
	meshSubset:           3,
	MeshService:          5,
	MeshExternalService:  6,
	MeshMultiZoneService: 7,
	meshServiceSubset:    8,
	MeshHTTPRoute:        9,
}

func (k TargetRefKind) Compare(o TargetRefKind) int {
	return order[k] - order[o]
}

func (k TargetRefKind) IsRealResource() bool {
	switch k {
	case meshSubset, meshServiceSubset:
		return false
	default:
		return true
	}
}

// These are the kinds that can be used in Kuma policies before support for
// actual resources (e.g., MeshExternalService, MeshMultiZoneService, and MeshService) was introduced.
func (k TargetRefKind) IsOldKind() bool {
	switch k {
	case Mesh, meshSubset, meshServiceSubset, MeshService, MeshHTTPRoute:
		return true
	default:
		return false
	}
}

func AllTargetRefKinds() []TargetRefKind {
	keys := util_maps.AllKeys(order)
	sort.Sort(TargetRefKindSlice(keys))
	return keys
}

type TargetRefKindSlice []TargetRefKind

func (x TargetRefKindSlice) Len() int           { return len(x) }
func (x TargetRefKindSlice) Less(i, j int) bool { return string(x[i]) < string(x[j]) }
func (x TargetRefKindSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

// TargetRef defines structure that allows attaching policy to various objects
type TargetRef struct {
	// This is needed to not sync policies with empty topLevelTarget ref to old zones that does not support it
	// This can be removed in 2.11.x
	UsesSyntacticSugar bool `json:"-"`

	// Kind of the referenced resource
	// +kubebuilder:validation:Enum=Mesh;MeshSubset;MeshService;MeshExternalService;MeshMultiZoneService;MeshServiceSubset;MeshHTTPRoute;Dataplane
	Kind TargetRefKind `json:"kind"`
	// Tags used to select a subset of proxies by tags. Can only be used with kinds
	// `MeshSubset` and `MeshServiceSubset`
	Tags *map[string]string `json:"tags,omitempty"`
	// Labels are used to select referenced real resources and to carry legacy
	// service identity when a common TargetRef must still target old
	// service-tag based paths.
	Labels *map[string]string `json:"labels,omitempty"`
	// SectionName is used to target specific section of resource.
	// For example, you can target port from MeshService.ports[] by its name. Only traffic to this port will be affected.
	SectionName *string `json:"sectionName,omitempty"`
}

func (t TargetRef) CompareDataplaneKind(other TargetRef) int {
	if t.Kind != Dataplane || other.Kind != Dataplane {
		return 0
	}
	if pointer.Deref(t.SectionName) != "" && pointer.Deref(other.SectionName) == "" {
		return 1
	}
	if pointer.Deref(t.SectionName) == "" && pointer.Deref(other.SectionName) != "" {
		return -1
	}
	return 0
}

// IncludesGateways reports whether a policy attached with this targetRef could
// apply to a Gateway-type dataplane (a delegated gateway is an ordinary
// Dataplane from the CP's perspective, not a distinct kind). Kind: Mesh (and
// the legacy MeshSubset) has no way to exclude gateways, so it always includes
// them; Kind: Dataplane never distinguishes gateways from any other dataplane,
// same as before proxyTypes existed (it was never a valid field on Kind:
// Dataplane); MeshHTTPRoute is always gateway-routing.
func IncludesGateways(ref TargetRef) bool {
	switch ref.Kind {
	case Mesh, meshSubset, MeshHTTPRoute:
		return true
	default:
		return false
	}
}

// +kubebuilder:validation:Enum=MeshOpenTelemetryBackend
type BackendResourceKind string

const (
	BackendResourceMeshOpenTelemetryBackend BackendResourceKind = "MeshOpenTelemetryBackend"
)

// BackendResourceRef is a reference to a backend resource within the same
// mesh. Used by observability policies to point at a MeshOpenTelemetryBackend
// via label matching. When multiple resources match, the oldest by creation time wins.
type BackendResourceRef struct {
	// Kind of the backend resource.
	Kind BackendResourceKind `json:"kind"`
	// Labels to match the referenced resource. When multiple resources match,
	// the oldest by creation time wins.
	Labels map[string]string `json:"labels,omitempty"`
}

// BackendRef defines where to forward traffic.
type BackendRef struct {
	// +kuma:nolint // https://github.com/kumahq/kuma/issues/14107
	TargetRef `json:","`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	// +kuma:nolint // https://github.com/kumahq/kuma/issues/14107
	Weight *uint `json:"weight,omitempty"`
	// Port is only supported when this ref refers to a real MeshService object
	Port *uint32 `json:"port,omitempty"`
}

func (b BackendRef) ReferencesRealObject() bool {
	switch b.Kind {
	case MeshService, MeshExternalService, MeshMultiZoneService:
		return true
	case meshServiceSubset:
		return false
	// empty targetRef should not be treated as real object
	case "":
		return false
	default:
		return true
	}
}

// MatchesHash is used to hash route matches to determine the origin resource
// for a ref
type MatchesHash string

type BackendRefHash string

func (b BackendRef) RealResourceSelector(defaultNamespace string) (map[string]string, string, bool) {
	if !b.ReferencesRealObject() {
		return nil, "", false
	}

	labels, sectionName, ok := realResourceSelector(b.TargetRef, defaultNamespace)
	if !ok {
		return nil, "", false
	}

	if port := pointer.Deref(b.Port); port > 0 && sectionName == "" {
		sectionName = fmt.Sprintf("%d", port)
	}

	return labels, sectionName, true
}

// Hash returns a hash of the BackendRef
func (in BackendRef) Hash() BackendRefHash {
	labels := pointer.Deref(in.Labels)
	sectionName := pointer.Deref(in.SectionName)
	if in.ReferencesRealObject() {
		if selectorLabels, selectorSectionName, ok := in.RealResourceSelector(""); ok {
			labels = selectorLabels
			sectionName = selectorSectionName
		}
	}

	keys := util_maps.SortedKeys(pointer.Deref(in.Tags))
	orderedTags := make([]string, 0, len(keys))
	for _, k := range keys {
		orderedTags = append(orderedTags, fmt.Sprintf("%s=%s", k, pointer.Deref(in.Tags)[k]))
	}

	keys = util_maps.SortedKeys(labels)
	orderedLabels := make([]string, 0, len(labels))
	for _, k := range keys {
		orderedLabels = append(orderedLabels, fmt.Sprintf("%s=%s", k, labels[k]))
	}

	return BackendRefHash(fmt.Sprintf(
		"%s/%s/%s/%d/%s",
		in.Kind,
		strings.Join(orderedTags, "/"),
		strings.Join(orderedLabels, "/"),
		pointer.DerefOr(in.Port, 0),
		sectionName,
	))
}

func realResourceSelector(ref TargetRef, defaultNamespace string) (map[string]string, string, bool) {
	if len(pointer.Deref(ref.Labels)) > 0 {
		labels := cloneStringMap(pointer.Deref(ref.Labels))
		switch ref.Kind {
		case MeshService, MeshExternalService, MeshMultiZoneService:
			if defaultNamespace != "" &&
				labels[mesh_proto.DisplayName] != "" &&
				labels[mesh_proto.KubeNamespaceTag] == "" {
				labels[mesh_proto.KubeNamespaceTag] = defaultNamespace
			}
		}
		return labels, pointer.Deref(ref.SectionName), true
	}
	return nil, "", false
}

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	maps.Copy(out, in)
	return out
}
