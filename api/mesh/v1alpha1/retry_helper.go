package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	proto_util "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *Retry) UnmarshalJSON(data []byte) error {
	return proto_util.FromJSON(data, m)
}

func (m *Retry) MarshalJSON() ([]byte, error) {
	return proto_util.ToJSON(m)
}
func (t *Retry) DeepCopyInto(out *Retry) {
	proto.Merge(out, t)
}
func (t *Retry) DeepCopy() *Retry {
	if t == nil {
		return nil
	}
	out := new(Retry)
	t.DeepCopyInto(out)
	return out
}
