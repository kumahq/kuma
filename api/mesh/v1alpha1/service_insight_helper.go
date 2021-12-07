package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	proto_util "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *ServiceInsight) UnmarshalJSON(data []byte) error {
	return proto_util.FromJSON(data, m)
}

func (m *ServiceInsight) MarshalJSON() ([]byte, error) {
	return proto_util.ToJSON(m)
}
func (t *ServiceInsight) DeepCopyInto(out *ServiceInsight) {
	proto.Merge(out, t)
}
func (t *ServiceInsight) DeepCopy() *ServiceInsight {
	if t == nil {
		return nil
	}
	out := new(ServiceInsight)
	t.DeepCopyInto(out)
	return out
}
