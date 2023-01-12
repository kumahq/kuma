// +kubebuilder:object:generate=true
package v1alpha1

// DataSource defines the source of bytes to use.
type DataSource struct {
	// Data source is a secret with given Secret key.
	Secret *string `json:"secret,omitempty"`
	// Data source is inline bytes.
	Inline *[]byte `json:"inline,omitempty"`
	// Data source is inline string`
	InlineString *string `json:"inlineString,omitempty"`
}
