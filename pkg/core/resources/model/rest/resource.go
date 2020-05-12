package rest

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/core/resources/model"
)

type ResourceMeta struct {
	Type             string    `json:"type"`
	Mesh             string    `json:"mesh,omitempty"`
	Name             string    `json:"name"`
	CreationTime     time.Time `json:"creationTime"`
	ModificationTime time.Time `json:"modificationTime"`
}

type Resource struct {
	Meta ResourceMeta
	Spec model.ResourceSpec
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
