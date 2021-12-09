package types

import (
	"bytes"
	"encoding/json"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

type InspectEntry struct {
	Type            string                                                `json:"type"`
	Name            string                                                `json:"name"`
	MatchedPolicies map[core_model.ResourceType][]core_model.ResourceSpec `json:"matchedPolicies"`
}

var _ json.Marshaler = &InspectEntry{}
var _ json.Unmarshaler = &InspectEntry{}

func (r *InspectEntry) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{}
	if r.Type != "" {
		result["type"] = r.Type
	}
	if r.Name != "" {
		result["name"] = r.Name
	}

	matchedPolicyMap := map[string]interface{}{}
	for resType, matchedPolicies := range r.MatchedPolicies {
		list := []interface{}{}
		for _, item := range matchedPolicies {
			var buf bytes.Buffer
			if err := (&jsonpb.Marshaler{}).Marshal(&buf, item); err != nil {
				return nil, err
			}
			out := map[string]interface{}{}
			if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
				return nil, err
			}
			list = append(list, out)
		}
		matchedPolicyMap[string(resType)] = list
	}
	if len(matchedPolicyMap) != 0 {
		result["matchedPolicies"] = matchedPolicyMap
	}
	return json.Marshal(result)
}

func (r *InspectEntry) UnmarshalJSON(data []byte) error {
	m := map[string]interface{}{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if m["name"] != nil {
		r.Name = m["name"].(string)
	}
	if m["type"] != nil {
		r.Type = m["type"].(string)
	}

	if matchedPoliciesRaw, ok := m["matchedPolicies"]; ok {
		r.MatchedPolicies = map[core_model.ResourceType][]core_model.ResourceSpec{}

		matchedPolicies, ok := matchedPoliciesRaw.(map[string]interface{})
		if !ok {
			return errors.New("MatchedPolicies is not a map[string]interface{}")
		}
		for resTypeRaw, matchedPolicyRaw := range matchedPolicies {
			var resType core_model.ResourceType
			if resTypeRaw == "" {
				return errors.New("MatchedPolicies key is empty")
			} else {
				resType = core_model.ResourceType(resTypeRaw)
			}

			items, ok := matchedPolicyRaw.([]interface{})
			if !ok {
				return errors.Errorf("MatchedPolicies[%s] is not a list", resType)
			}
			for _, item := range items {
				res, err := registry.Global().NewObject(resType)
				if err != nil {
					return err
				}
				mItem, err := json.Marshal(item)
				if err != nil {
					return err
				}
				if err := (&jsonpb.Unmarshaler{}).Unmarshal(bytes.NewReader(mItem), res.GetSpec()); err != nil {
					return err
				}
				r.MatchedPolicies[resType] = append(r.MatchedPolicies[resType], res.GetSpec())
			}
		}
	}
	return nil
}
