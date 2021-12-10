package v1alpha1

import (
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *Retry) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *Retry) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *Retry) DeepCopyInto(out *Retry) {
	util_proto.Merge(out, t)
}
func (t *Retry) DeepCopy() *Retry {
	if t == nil {
		return nil
	}
	out := new(Retry)
	t.DeepCopyInto(out)
	return out
}
