package v1alpha1

// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=256
// +kubebuilder:validation:Pattern=`^[a-z0-9!#$%&'*+\-.^_\x60|~]+$`
type HeaderName string

type HeaderValue string
