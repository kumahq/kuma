package v1alpha1

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v2/pkg/core/resources/sni"
)

type Resource struct {
	ResourceMeta
	Spec   core_model.ResourceSpec   `json:"spec,omitempty"`
	Status core_model.ResourceStatus `json:"status,omitempty"`
}

var _ json.Marshaler = (*Resource)(nil)

func (r *Resource) MarshalJSON() ([]byte, error) {
	var specJSON json.RawMessage
	if r.Spec != nil {
		b, err := core_model.ToJSON(r.Spec)
		if err != nil {
			return nil, err
		}
		specJSON = b
	}

	var statusJSON json.RawMessage
	if r.Status != nil {
		b, err := core_model.ToJSON(r.Status)
		if err != nil {
			return nil, err
		}
		statusJSON = b
	}

	var kriStr string
	var snis []SNIEntry
	if id, _ := kri.FromResourceMetaE(r.ResourceMeta, r.Type); !id.IsEmpty() {
		kriStr = id.String()
		snis = computeSNIs(id, r.Spec)
	}

	// Explicit struct with all fields in desired order
	// Cannot embed ResourceMeta as it has MarshalJSON which would override this struct's marshaling
	aux := &struct {
		Type             string            `json:"type"`
		Mesh             string            `json:"mesh,omitempty"`
		Name             string            `json:"name"`
		CreationTime     time.Time         `json:"creationTime"`
		ModificationTime time.Time         `json:"modificationTime"`
		Labels           map[string]string `json:"labels,omitempty"`
		KRI              string            `json:"kri,omitempty"`
		SNIs             []SNIEntry        `json:"snis,omitempty"`
		Spec             json.RawMessage   `json:"spec,omitempty"`
		Status           json.RawMessage   `json:"status,omitempty"`
	}{
		Type:             r.Type,
		Mesh:             r.Mesh,
		Name:             r.Name,
		CreationTime:     r.CreationTime,
		ModificationTime: r.ModificationTime,
		Labels:           r.Labels,
		KRI:              kriStr,
		SNIs:             snis,
		Spec:             specJSON,
		Status:           statusJSON,
	}

	return json.Marshal(aux)
}

// SNIEntry pairs a destination port with the SNI advertised by xDS for that
// port. Section is the section name fed into the trailing segment of the
// SNI (typically the port name, falling back to the stringified port).
type SNIEntry struct {
	Port int32  `json:"port"`
	SNI  string `json:"sni"`
}

// computeSNIs returns the list of SNIs for the resource identified by id,
// derived from the sections contributed by spec. Entries are sorted by port
// ascending so the output is deterministic. Returns nil for resources that
// are not destinations or whose spec does not implement sni.SectionLister.
func computeSNIs(id kri.Identifier, spec core_model.ResourceSpec) []SNIEntry {
	desc, err := registry.Global().DescriptorFor(id.ResourceType)
	if err != nil || !desc.IsDestination {
		return nil
	}
	lister, ok := spec.(sni.SectionLister)
	if !ok {
		return nil
	}
	sections := lister.SNIs()
	if len(sections) == 0 {
		return nil
	}
	out := make([]SNIEntry, 0, len(sections))
	for _, s := range sections {
		v, err := sni.FromKRIE(kri.WithSectionName(id, s.SectionName))
		if err != nil {
			continue
		}
		out = append(out, SNIEntry{Port: s.Port, SNI: v})
	}
	if len(out) == 0 {
		return nil
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Port < out[j].Port })
	return out
}

func (r *Resource) GetMeta() ResourceMeta {
	if r == nil {
		return ResourceMeta{}
	}
	return r.ResourceMeta
}

func (r *Resource) GetSpec() core_model.ResourceSpec {
	if r == nil {
		return nil
	}
	return r.Spec
}

func (r *Resource) GetStatus() core_model.ResourceStatus {
	if r == nil {
		return nil
	}
	return r.Status
}
