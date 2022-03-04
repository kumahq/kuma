package types

import (
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

func NewPolicyInspectEntryList() *PolicyInspectEntryList {
	return &PolicyInspectEntryList{
		Total: 0,
		Items: []*PolicyInspectEntry{},
	}
}

type MatchedPolicies map[core_model.ResourceType][]*rest.Resource

type DataplaneInspectEntry struct {
	AttachmentEntry
	MatchedPolicies MatchedPolicies `json:"matchedPolicies"`
}

type DataplaneInspectEntryList struct {
	Total uint32                   `json:"total"`
	Kind  string                   `json:"kind"`
	Items []*DataplaneInspectEntry `json:"items"`
}

func NewDataplaneInspectEntryList() *DataplaneInspectEntryList {
	return &DataplaneInspectEntryList{
		Total: 0,
		Kind:  "SidecarDataplane",
		Items: []*DataplaneInspectEntry{},
	}
}
