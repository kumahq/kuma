// +kubebuilder:object:generate=true
package v1alpha1

type LabelSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type Selector struct {
	MeshService LabelSelector `json:"meshService,omitempty"`
}

// HostnameGenerator
// +kuma:policy:is_policy=false
type HostnameGenerator struct {
	Selector Selector `json:"selector,omitempty"`
	Template string   `json:"template,omitempty"`
}
