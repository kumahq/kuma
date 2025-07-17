// +kubebuilder:object:generate=true
package v1alpha1

import "github.com/kumahq/kuma/pkg/util/pointer"

type LabelSelector struct {
	MatchLabels *map[string]string `json:"matchLabels,omitempty"`
}

func (s LabelSelector) Matches(labels map[string]string) bool {
	for tag, matchValue := range pointer.Deref(s.MatchLabels) {
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
