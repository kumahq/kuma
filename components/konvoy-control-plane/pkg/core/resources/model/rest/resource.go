package rest

import (
	"bytes"
	"encoding/json"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/gogo/protobuf/jsonpb"
)

type ResourceMeta struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Mesh string `json:"mesh"`
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
