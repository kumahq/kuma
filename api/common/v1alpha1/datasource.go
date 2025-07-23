// +kubebuilder:object:generate=true
package v1alpha1

import (
	"google.golang.org/protobuf/types/known/wrapperspb"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
)

// Deprecated: use api/common/v1alpha1/datasource/datasource.go
// DataSource defines the source of bytes to use.
type DataSource struct {
	// Data source is a secret with given Secret key.
	Secret *string `json:"secret,omitempty"`
	// Data source is inline bytes.
	Inline *[]byte `json:"inline,omitempty"`
	// Data source is inline string`
	InlineString *string `json:"inlineString,omitempty"`
}

func (ds *DataSource) ConvertToProto() *system_proto.DataSource {
	switch {
	case ds.Secret != nil:
		return &system_proto.DataSource{Type: &system_proto.DataSource_Secret{Secret: *ds.Secret}}
	case ds.Inline != nil:
		return &system_proto.DataSource{Type: &system_proto.DataSource_Inline{Inline: &wrapperspb.BytesValue{Value: *ds.Inline}}}
	case ds.InlineString != nil:
		return &system_proto.DataSource{Type: &system_proto.DataSource_InlineString{InlineString: *ds.InlineString}}
	}
	return nil
}
