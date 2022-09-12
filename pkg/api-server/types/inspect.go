package types

import (
	"encoding/json"

	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
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

func ResourceKeyEntryFromModelKey(key core_model.ResourceKey) ResourceKeyEntry {
	return ResourceKeyEntry{
		Mesh: key.Mesh,
		Name: key.Name,
	}
}

type PolicyInspectEntryKind interface {
	policyInspectEntry()
}

type PolicyInspectEntry struct {
	PolicyInspectEntryKind
}

func NewPolicyInspectEntry(k PolicyInspectEntryKind) PolicyInspectEntry {
	return PolicyInspectEntry{PolicyInspectEntryKind: k}
}

func (w *PolicyInspectEntry) UnmarshalJSON(data []byte) error {
	i := KindTag{}
	if err := json.Unmarshal(data, &i); err != nil {
		return errors.Wrap(err, `unable to find "kind"`)
	}
	var entry PolicyInspectEntryKind
	switch i.Kind {
	// We treat a non-kinded entry as a SidecarDataplane for backwards
	// compatibility
	case SidecarDataplane, "":
		entry = &PolicyInspectSidecarEntry{}
	case GatewayDataplane:
		entry = &PolicyInspectGatewayEntry{}
	default:
		return errors.Errorf("invalid PolicyInspectEntry kind %q", i.Kind)
	}
	if err := json.Unmarshal(data, entry); err != nil {
		return errors.Wrapf(err, "unable to parse PolicyInspectEntry of kind %q", i.Kind)
	}
	w.PolicyInspectEntryKind = entry
	return nil
}

type PolicyInspectSidecarEntry struct {
	DataplaneKey ResourceKeyEntry  `json:"dataplane"`
	Attachments  []AttachmentEntry `json:"attachments"`
}

const SidecarDataplane = "SidecarDataplane"
const GatewayDataplane = "MeshGatewayDataplane"

type KindTag struct {
	Kind string `json:"kind"`
}

func (e PolicyInspectEntry) MarshalJSON() ([]byte, error) {
	switch concrete := e.PolicyInspectEntryKind.(type) {
	case *PolicyInspectSidecarEntry:
		return json.Marshal(struct {
			KindTag
			*PolicyInspectSidecarEntry
		}{
			KindTag:                   KindTag{SidecarDataplane},
			PolicyInspectSidecarEntry: concrete,
		})
	case *PolicyInspectGatewayEntry:
		return json.Marshal(struct {
			KindTag
			*PolicyInspectGatewayEntry
		}{
			KindTag:                   KindTag{GatewayDataplane},
			PolicyInspectGatewayEntry: concrete,
		})
	}
	panic("internal error")
}

func (*PolicyInspectSidecarEntry) policyInspectEntry() {
}

func NewPolicyInspectSidecarEntry(key ResourceKeyEntry) PolicyInspectSidecarEntry {
	return PolicyInspectSidecarEntry{
		DataplaneKey: key,
	}
}

type GatewayDataplanesInspectEntry struct {
	DataplaneKey ResourceKeyEntry `json:"dataplane"`
}

type GatewayDataplanesInspectEntryList struct {
	Total uint32                          `json:"total"`
	Items []GatewayDataplanesInspectEntry `json:"items"`
}

func NewGatewayDataplanesInspectResult() *GatewayDataplanesInspectEntryList {
	return &GatewayDataplanesInspectEntryList{
		Total: 0,
		Items: []GatewayDataplanesInspectEntry{},
	}
}

type PolicyInspectEntryList struct {
	Total uint32               `json:"total"`
	Items []PolicyInspectEntry `json:"items"`
}

func NewPolicyInspectEntryList() *PolicyInspectEntryList {
	return &PolicyInspectEntryList{
		Total: 0,
		Items: []PolicyInspectEntry{},
	}
}

type MatchedPolicies map[core_model.ResourceType][]*unversioned.Resource

type DataplaneInspectResponseKind interface {
	dataplaneInspectEntry()
}

type DataplaneInspectResponse struct {
	DataplaneInspectResponseKind
}

func NewDataplaneInspectResponse(k DataplaneInspectResponseKind) DataplaneInspectResponse {
	return DataplaneInspectResponse{
		DataplaneInspectResponseKind: k,
	}
}

func (e DataplaneInspectResponse) MarshalJSON() ([]byte, error) {
	switch concrete := e.DataplaneInspectResponseKind.(type) {
	case *DataplaneInspectEntryList:
		return json.Marshal(struct {
			KindTag
			*DataplaneInspectEntryList
		}{
			KindTag:                   KindTag{SidecarDataplane},
			DataplaneInspectEntryList: concrete,
		})
	case *GatewayDataplaneInspectResult:
		return json.Marshal(struct {
			KindTag
			*GatewayDataplaneInspectResult
		}{
			KindTag:                       KindTag{GatewayDataplane},
			GatewayDataplaneInspectResult: concrete,
		})
	}
	panic("internal error")
}

func (w *DataplaneInspectResponse) UnmarshalJSON(data []byte) error {
	i := KindTag{}
	if err := json.Unmarshal(data, &i); err != nil {
		return errors.Wrap(err, `unable to find "kind"`)
	}
	var entry DataplaneInspectResponseKind
	switch i.Kind {
	case SidecarDataplane, "":
		entry = &DataplaneInspectEntryList{}
	case GatewayDataplane:
		entry = &GatewayDataplaneInspectResult{}
	default:
		return errors.Errorf("invalid DataplaneInspectResponse kind %q", i.Kind)
	}
	if err := json.Unmarshal(data, entry); err != nil {
		return errors.Wrapf(err, "unable to parse DataplaneInspectResponse of kind %q", i.Kind)
	}
	w.DataplaneInspectResponseKind = entry
	return nil
}

type DataplaneInspectEntry struct {
	AttachmentEntry
	MatchedPolicies MatchedPolicies `json:"matchedPolicies"`
}

type DataplaneInspectEntryList struct {
	Total uint32                   `json:"total"`
	Items []*DataplaneInspectEntry `json:"items"`
}

func NewDataplaneInspectEntryList() *DataplaneInspectEntryList {
	return &DataplaneInspectEntryList{
		Total: 0,
		Items: []*DataplaneInspectEntry{},
	}
}

func (*DataplaneInspectEntryList) dataplaneInspectEntry() {
}
