package rest

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

type ResourceMeta struct {
	Type             string    `json:"type"`
	Mesh             string    `json:"mesh,omitempty"`
	Name             string    `json:"name"`
	CreationTime     time.Time `json:"creationTime"`
	ModificationTime time.Time `json:"modificationTime"`
}

func (r *ResourceMeta) GetName() string {
	return r.Name
}

func (r *ResourceMeta) GetNameExtensions() model.ResourceNameExtensions {
	return model.ResourceNameExtensionsUnsupported
}

func (r *ResourceMeta) GetVersion() string {
	return ""
}

func (r *ResourceMeta) GetMesh() string {
	return r.Mesh
}

func (r *ResourceMeta) GetCreationTime() time.Time {
	return r.CreationTime
}

func (r *ResourceMeta) GetModificationTime() time.Time {
	return r.ModificationTime
}

var _ model.ResourceMeta = &ResourceMeta{}

type Resource struct {
	Meta ResourceMeta
	Spec model.ResourceSpec
}

// NewFromModel create a REST Resource from the given model Resource.
func NewFromModel(m model.Resource) *Resource {
	if m == nil {
		return nil
	}

	meta := m.GetMeta()
	return &Resource{
		Meta: ResourceMeta{
			Type:             string(m.Descriptor().Name),
			Mesh:             meta.GetMesh(),
			Name:             meta.GetName(),
			CreationTime:     meta.GetCreationTime(),
			ModificationTime: meta.GetModificationTime(),
		},
		Spec: m.GetSpec(),
	}
}

type ResourceList struct {
	Total uint32      `json:"total"`
	Items []*Resource `json:"items"`
	Next  *string     `json:"next"`
}

var _ json.Marshaler = &Resource{}
var _ json.Unmarshaler = &Resource{}

func (r *Resource) MarshalJSON() ([]byte, error) {
	var specBytes []byte
	if r.Spec != nil {
		var buf bytes.Buffer
		if err := (&jsonpb.Marshaler{}).Marshal(&buf, r.Spec); err != nil {
			return nil, err
		}
		specBytes = buf.Bytes()
	}

	metaJSON, err := json.Marshal(r.Meta)
	if err != nil {
		return nil, err
	}

	if len(specBytes) == 0 || string(specBytes) == "{}" { // spec is nil or empty
		return metaJSON, nil
	} else {
		// remove the } of meta JSON, { of spec JSON and join it by ,
		return append(append(metaJSON[:len(metaJSON)-1], byte(',')), specBytes[1:]...), nil
	}
}

func (r *Resource) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.Meta); err != nil {
		return err
	}
	if r.Spec == nil {
		newR, err := registry.Global().NewObject(model.ResourceType(r.Meta.Type))
		if err != nil {
			return err
		}
		r.Spec = newR.GetSpec()
	}
	if err := (&jsonpb.Unmarshaler{AllowUnknownFields: true}).Unmarshal(bytes.NewReader(data), r.Spec); err != nil {
		return err
	}
	return nil
}

type ResourceListReceiver struct {
	ResourceList
	NewResource func() model.Resource
}

var _ json.Unmarshaler = &ResourceListReceiver{}

func (rec *ResourceListReceiver) UnmarshalJSON(data []byte) error {
	if rec.NewResource == nil {
		return errors.Errorf("NewResource must not be nil")
	}
	type List struct {
		Total uint32             `json:"total"`
		Items []*json.RawMessage `json:"items"`
		Next  *string            `json:"next"`
	}
	list := List{}
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	rec.ResourceList.Total = list.Total
	rec.ResourceList.Items = make([]*Resource, len(list.Items))
	for i, li := range list.Items {
		b, err := json.Marshal(li)
		if err != nil {
			return err
		}
		r := &Resource{}
		if err := json.Unmarshal(b, &r.Meta); err != nil {
			return err
		}
		r.Spec = rec.NewResource().GetSpec()
		if err := json.Unmarshal(b, r); err != nil {
			return err
		}
		rec.ResourceList.Items[i] = r
	}
	rec.ResourceList.Next = list.Next
	return nil
}

type ByMeta []*Resource

func (a ByMeta) Len() int { return len(a) }

func (a ByMeta) Less(i, j int) bool {
	if a[i].Meta.Mesh == a[j].Meta.Mesh {
		return a[i].Meta.Name < a[j].Meta.Name
	}
	return a[i].Meta.Mesh < a[j].Meta.Mesh
}

func (a ByMeta) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
