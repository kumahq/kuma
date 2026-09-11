// +kubebuilder:object:generate=true
package v1alpha1

// DoNothingResource
// +kuma:policy:skip_registration=true
// +kuma:policy:is_policy=false
type DoNothingResource struct{}
