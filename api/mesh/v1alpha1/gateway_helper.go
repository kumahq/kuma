package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	proto_util "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *Gateway) UnmarshalJSON(data []byte) error {
	return proto_util.FromJSON(data, m)
}

func (m *Gateway) MarshalJSON() ([]byte, error) {
	return proto_util.ToJSON(m)
}
func (t *Gateway) DeepCopyInto(out *Gateway) {
	proto.Merge(out, t)
}
func (t *Gateway) DeepCopy() *Gateway {
	if t == nil {
		return nil
	}
	out := new(Gateway)
	t.DeepCopyInto(out)
	return out
}
