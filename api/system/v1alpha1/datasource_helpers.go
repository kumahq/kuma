package v1alpha1

import util_proto "github.com/kumahq/kuma/pkg/util/proto"

func (ds *DataSource) MaskInlineDatasource() *DataSource {
	if ds == nil {
		return nil
	}
	if ds.GetInline().String() != "" {
		return &DataSource{
			Type: &DataSource_Inline{Inline: util_proto.Bytes([]byte("***"))},
		}
	}
	if ds.GetInlineString() != "" {
		return &DataSource{
			Type: &DataSource_InlineString{InlineString: "***"},
		}
	}
	return nil
}
