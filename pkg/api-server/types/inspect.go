package types

import (
	"bytes"
	"encoding/json"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

type InspectEntry struct {
	Type            string                                            `json:"type"`
	Name            string                                            `json:"name"`
	MatchedPolicies map[core_model.ResourceType][]core_model.Resource `json:"matchedPolicies"`
}

type intermediateInspectEntry struct {
	// Intermediate representation is needed to preserve the order of the fields.
	// If use 'map[string]interface{}' as intermediate representation then the fields
	// will be sorted lexicographically.
	Type            string                             `json:"type"`
	Name            string                             `json:"name"`
	MatchedPolicies map[string][]*intermediateResource `json:"matchedPolicies"`
}

type intermediateResource struct {
	Meta rest.ResourceMeta      `json:"meta"`
	Spec map[string]interface{} `json:"spec"`
}

var _ json.Marshaler = &InspectEntry{}
var _ json.Unmarshaler = &InspectEntry{}

func (r *InspectEntry) MarshalJSON() ([]byte, error) {
	intermediate, err := r.toIntermediate()
	if err != nil {
		return nil, err
	}
	return json.Marshal(intermediate)
}

func (r *InspectEntry) toIntermediate() (*intermediateInspectEntry, error) {
	result := &intermediateInspectEntry{
		Type: r.Type,
		Name: r.Name,
	}
	matchedPolicies, err := r.matchedPoliciesMapToIntermediate()
	if err != nil {
		return nil, err
	}
	result.MatchedPolicies = matchedPolicies
	return result, nil
}

func (r *InspectEntry) matchedPoliciesMapToIntermediate() (map[string][]*intermediateResource, error) {
	matchedPolicyMap := map[string][]*intermediateResource{}
	for resType, matchedPolicies := range r.MatchedPolicies {
		resources, err := r.resourcesToIntermediate(matchedPolicies)
		if err != nil {
			return nil, err
		}
		matchedPolicyMap[string(resType)] = resources
	}
	if len(matchedPolicyMap) == 0 {
		return nil, nil
	}
	return matchedPolicyMap, nil
}

func (r *InspectEntry) resourcesToIntermediate(rs []core_model.Resource) ([]*intermediateResource, error) {
	list := []*intermediateResource{}
	for _, item := range rs {
		intermRes, err := r.resourceToIntermediate(item)
		if err != nil {
			return nil, err
		}
		list = append(list, intermRes)
	}
	return list, nil
}

func (r *InspectEntry) resourceToIntermediate(resource core_model.Resource) (*intermediateResource, error) {
	intermediate := &intermediateResource{
		Meta: rest.ResourceMeta{
			Type:             string(resource.Descriptor().Name),
			Mesh:             resource.GetMeta().GetMesh(),
			Name:             resource.GetMeta().GetName(),
			CreationTime:     resource.GetMeta().GetCreationTime(),
			ModificationTime: resource.GetMeta().GetModificationTime(),
		},
	}
	var buf bytes.Buffer
	if err := (&jsonpb.Marshaler{}).Marshal(&buf, resource.GetSpec()); err != nil {
		return nil, err
	}
	spec := map[string]interface{}{}
	if err := json.Unmarshal(buf.Bytes(), &spec); err != nil {
		return nil, err
	}
	intermediate.Spec = spec
	return intermediate, nil
}

func (r *InspectEntry) UnmarshalJSON(data []byte) error {
	intermediate := intermediateInspectEntry{}
	if err := json.Unmarshal(data, &intermediate); err != nil {
		return err
	}

	if intermediate.MatchedPolicies != nil {
		r.MatchedPolicies = map[core_model.ResourceType][]core_model.Resource{}

		for resTypeRaw, intermResources := range intermediate.MatchedPolicies {
			var resType core_model.ResourceType
			if resTypeRaw == "" {
				return errors.New("MatchedPolicies key is empty")
			} else {
				resType = core_model.ResourceType(resTypeRaw)
			}

			for _, item := range intermResources {
				res, err := registry.Global().NewObject(resType)
				if err != nil {
					return err
				}
				mItem, err := json.Marshal(item.Spec)
				if err != nil {
					return err
				}
				if err := (&jsonpb.Unmarshaler{}).Unmarshal(bytes.NewReader(mItem), res.GetSpec()); err != nil {
					return err
				}
				res.SetMeta(&item.Meta)
				r.MatchedPolicies[resType] = append(r.MatchedPolicies[resType], res)
			}
		}
	}

	r.Name = intermediate.Name
	r.Type = intermediate.Type

	return nil
}
