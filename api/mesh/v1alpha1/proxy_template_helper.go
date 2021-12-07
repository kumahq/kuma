package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	proto_util "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *ProxyTemplate) UnmarshalJSON(data []byte) error {
	return proto_util.FromJSON(data, m)
}

func (m *ProxyTemplate) MarshalJSON() ([]byte, error) {
	return proto_util.ToJSON(m)
}
func (t *ProxyTemplate) DeepCopyInto(out *ProxyTemplate) {
	proto.Merge(out, t)
}
func (t *ProxyTemplate) DeepCopy() *ProxyTemplate {
	if t == nil {
		return nil
	}
	out := new(ProxyTemplate)
	t.DeepCopyInto(out)
	return out
}
