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
	Type            string          `json:"type"`
	Name            string          `json:"name"`
	MatchedPolicies []MatchedPolicy `json:"matchedPolicies"`
}

type MatchedPolicy struct {
	ResourceType core_model.ResourceType   `json:"resourceType"`
	Items        []core_model.ResourceSpec `json:"items"`
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
	plist := []interface{}{}
	for _, matchedPolicy := range r.MatchedPolicies {
		matchedPolicyMap := map[string]interface{}{}
		if matchedPolicy.ResourceType != "" {
			matchedPolicyMap["resourceType"] = matchedPolicy.ResourceType
		}

		list := []interface{}{}
		for _, item := range matchedPolicy.Items {
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
		matchedPolicyMap["items"] = list
		plist = append(plist, matchedPolicyMap)
	}
	if len(plist) != 0 {
		result["matchedPolicies"] = plist
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
		matchedPolicies, ok := matchedPoliciesRaw.([]interface{})
		if !ok {
			return errors.New("MatchedPolicies is not a list")
		}
		for i, matchedPolicyRaw := range matchedPolicies {
			unmarshalled := MatchedPolicy{}

			matchedPolicy, ok := matchedPolicyRaw.(map[string]interface{})
			if !ok {
				return errors.Errorf("MatchedPolicies[%d] is not a map[string]interface{}", i)
			}
			if matchedPolicy["resourceType"] != nil {
				unmarshalled.ResourceType = core_model.ResourceType(matchedPolicy["resourceType"].(string))
			}

			if itemsRaw, ok := matchedPolicy["items"]; ok {
				items, ok := itemsRaw.([]interface{})
				if !ok {
					return errors.Errorf("MatchedPolicies[%d].Items is not a list", i)
				}
				for _, item := range items {
					res, err := registry.Global().NewObject(unmarshalled.ResourceType)
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
					unmarshalled.Items = append(unmarshalled.Items, res.GetSpec())
				}
			}
			r.MatchedPolicies = append(r.MatchedPolicies, unmarshalled)
		}
	}
	return nil
}
