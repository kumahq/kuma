package v1alpha1

import "github.com/kumahq/kuma/v3/pkg/util/pointer"

func (ds *DataSource) MaskInlineDatasource() *DataSource {
	if ds == nil {
		return nil
	}
	if len(ds.GetInline().GetValue()) > 0 {
		return &DataSource{Inline: Bytes([]byte("***"))}
	}
	if ds.GetInlineString() != "" {
		return &DataSource{InlineString: pointer.To("***")}
	}
	return nil
}
