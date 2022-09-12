package rest

import (
	"encoding/json"

	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type ResourceList struct {
	Total uint32     `json:"total"`
	Items []Resource `json:"items"`
	Next  *string    `json:"next"`
}

type ResourceListReceiver struct {
	ResourceList
	NewResource func() core_model.Resource
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
	rec.ResourceList.Items = make([]Resource, len(list.Items))
	for i, li := range list.Items {
		b, err := json.Marshal(li)
		if err != nil {
			return err
		}

		restResource := From.Resource(rec.NewResource())
		if err := json.Unmarshal(b, restResource); err != nil {
			return err
		}

		rec.ResourceList.Items[i] = restResource
	}
	rec.ResourceList.Next = list.Next
	return nil
}
