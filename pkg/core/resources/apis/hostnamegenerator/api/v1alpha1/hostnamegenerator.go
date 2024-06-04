// +kubebuilder:object:generate=true
package v1alpha1

type LabelSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type Selector struct {
	MeshService LabelSelector `json:"meshService,omitempty"`
}

func (s LabelSelector) Matches(labels map[string]string) bool {
	for tag, matchValue := range s.MatchLabels {
		labelValue, exist := labels[tag]
		if !exist {
			return false
		}
		if matchValue != labelValue {
			return false
		}
	}
	return true
}

// HostnameGenerator
// +kuma:policy:is_policy=false
type HostnameGenerator struct {
	Selector Selector `json:"selector,omitempty"`
	Template string   `json:"template,omitempty"`
}
