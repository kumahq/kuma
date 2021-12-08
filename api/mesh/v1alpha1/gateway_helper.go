package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *Gateway) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *Gateway) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
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
