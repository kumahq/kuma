package types

import (
	"encoding/json"

	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
)

type AttachmentEntry struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Service string `json:"service"`
}

type ResourceKeyEntry struct {
	Mesh string `json:"mesh"`
	Name string `json:"name"`
}

type PolicyInspectEntry struct {
	DataplaneKey ResourceKeyEntry  `json:"dataplane"`
	Attachments  []AttachmentEntry `json:"attachments"`
}

type PolicyInspectEntryList struct {
	Total uint32                `json:"total"`
	Items []*PolicyInspectEntry `json:"items"`
}

type DataplaneInspectEntry struct {
	AttachmentEntry
	MatchedPolicies map[core_model.ResourceType][]*rest.Resource `json:"matchedPolicies"`
}

type DataplaneInspectEntryList struct {
	Total uint32                   `json:"total"`
	Items []*DataplaneInspectEntry `json:"items"`
}

type DataplaneInspectEntryReceiver struct {
	DataplaneInspectEntry
	NewResource func(resourceType core_model.ResourceType) (core_model.Resource, error)
}

var _ json.Unmarshaler = &DataplaneInspectEntryReceiver{}

func (rec *DataplaneInspectEntryReceiver) UnmarshalJSON(bytes []byte) error {
	if rec.NewResource == nil {
		return errors.Errorf("NewResource must not be nil")
	}

	type intermediate struct {
		Type            string                                        `json:"type"`
		Name            string                                        `json:"name"`
		Service         string                                        `json:"service"`
		MatchedPolicies map[core_model.ResourceType][]json.RawMessage `json:"matchedPolicies"`
	}
	inter := &intermediate{}

	if err := json.Unmarshal(bytes, inter); err != nil {
		return err
	}
	rec.Type = inter.Type
	rec.Name = inter.Name
	rec.Service = inter.Service
	rec.MatchedPolicies = map[core_model.ResourceType][]*rest.Resource{}

	for typ, rawList := range inter.MatchedPolicies {
		for _, rawItem := range rawList {
			res, err := rec.NewResource(typ)
			if err != nil {
				return err
			}
			restRes := &rest.Resource{
				Spec: res.GetSpec(),
			}
			if err := json.Unmarshal(rawItem, restRes); err != nil {
				return err
			}
			rec.MatchedPolicies[typ] = append(rec.MatchedPolicies[typ], restRes)
		}
	}

	return nil
}
