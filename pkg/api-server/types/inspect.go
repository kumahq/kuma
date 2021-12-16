package types

import (
	"bytes"
	"encoding/json"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

type PolicyInspectEntry struct {
	DataplaneKey ResourceKeyEntry  `json:"dataplane"`
	Attachments  []AttachmentEntry `json:"attachments"`
}

type DataplaneInspectEntry struct {
	AttachmentEntry
	MatchedPolicies map[core_model.ResourceType][]core_model.ResourceSpec `json:"matchedPolicies"`
}

type AttachmentEntry struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type ResourceKeyEntry struct {
	Mesh string `json:"mesh"`
	Name string `json:"name"`
}

type intermediateDataplaneInspectEntry struct {
	// Intermediate representation is needed to preserve the order of the fields.
	// If use 'map[string]interface{}' as intermediate representation then the fields
	// will be sorted lexicographically.
	Type            string                 `json:"type"`
	Name            string                 `json:"name"`
	MatchedPolicies map[string]interface{} `json:"matchedPolicies"`
}

var _ json.Marshaler = &DataplaneInspectEntry{}
var _ json.Unmarshaler = &DataplaneInspectEntry{}

func (r *DataplaneInspectEntry) MarshalJSON() ([]byte, error) {
	result := intermediateDataplaneInspectEntry{}

	result.Type = r.Type
	result.Name = r.Name

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
		result.MatchedPolicies = matchedPolicyMap
	}
	return json.Marshal(result)
}

func (r *DataplaneInspectEntry) UnmarshalJSON(data []byte) error {
	intermediate := intermediateDataplaneInspectEntry{}
	if err := json.Unmarshal(data, &intermediate); err != nil {
		return err
	}

	if intermediate.MatchedPolicies != nil {
		r.MatchedPolicies = map[core_model.ResourceType][]core_model.ResourceSpec{}

		for resTypeRaw, matchedPolicyRaw := range intermediate.MatchedPolicies {
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

	r.Name = intermediate.Name
	r.Type = intermediate.Type

	return nil
}
