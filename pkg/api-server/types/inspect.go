package types

import (
	"encoding/json"

	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
)

type InspectEntry struct {
	Type            string                                       `json:"type"`
	Name            string                                       `json:"name"`
	MatchedPolicies map[core_model.ResourceType][]*rest.Resource `json:"matchedPolicies"`
}

type InspectEntryReceiver struct {
	InspectEntry
	NewResource func(resourceType core_model.ResourceType) (core_model.Resource, error)
}

var _ json.Unmarshaler = &InspectEntryReceiver{}

func (rec *InspectEntryReceiver) UnmarshalJSON(bytes []byte) error {
	if rec.NewResource == nil {
		return errors.Errorf("NewResource must not be nil")
	}

	type intermediate struct {
		Type            string                                        `json:"type"`
		Name            string                                        `json:"name"`
		MatchedPolicies map[core_model.ResourceType][]json.RawMessage `json:"matchedPolicies"`
	}
	inter := &intermediate{}

	if err := json.Unmarshal(bytes, inter); err != nil {
		return err
	}
	rec.Type = inter.Type
	rec.Name = inter.Name
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
