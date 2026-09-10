package types

type AttachmentEntry struct {
	Type    string `json:"type"`
	Name    string `json:"name,omitempty"`
	Service string `json:"service,omitempty"`
}

type ResourceKeyEntry struct {
	Mesh string `json:"mesh"`
	Name string `json:"name"`
}

type PolicyInspectEntryKind interface {
	policyInspectEntry()
}

type PolicyInspectSidecarEntry struct {
	DataplaneKey ResourceKeyEntry  `json:"dataplane"`
	Attachments  []AttachmentEntry `json:"attachments"`
}

const (
	SidecarDataplane = "SidecarDataplane"
	GatewayDataplane = "MeshGatewayDataplane"
)

type KindTag struct {
	Kind string `json:"kind"`
}

func (*PolicyInspectSidecarEntry) policyInspectEntry() {
}
