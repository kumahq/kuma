package rest

import (
	"bytes"
	"encoding/json"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/gogo/protobuf/jsonpb"

	"github.com/pkg/errors"
)

type ResourceMeta struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Mesh string `json:"mesh,omitempty"`
}

type Resource struct {
	Meta ResourceMeta
	Spec model.ResourceSpec
}

type ResourceList struct {
	Items []*Resource `json:"items"`
}

var _ json.Marshaler = &Resource{}
var _ json.Unmarshaler = &Resource{}

func (r *Resource) MarshalJSON() ([]byte, error) {
	meta, err := json.Marshal(&r.Meta)
	if err != nil {
		return nil, err
	}
	if r.Spec == nil {
		return meta, nil
	}

	var buf bytes.Buffer
	if err := (&jsonpb.Marshaler{}).Marshal(&buf, r.Spec); err != nil {
		return nil, err
	}
	spec := buf.Bytes()

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(meta, &obj); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(spec, &obj); err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

func (r *Resource) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.Meta); err != nil {
		return err
	}
	if r.Spec == nil {
		return nil
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
		Items []*json.RawMessage `json:"items"`
	}
	list := List{}
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
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
	return nil
}
