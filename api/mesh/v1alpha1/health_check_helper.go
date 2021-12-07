package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	proto_util "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *HealthCheck) UnmarshalJSON(data []byte) error {
	return proto_util.FromJSON(data, m)
}

func (m *HealthCheck) MarshalJSON() ([]byte, error) {
	return proto_util.ToJSON(m)
}
func (t *HealthCheck) DeepCopyInto(out *HealthCheck) {
	proto.Merge(out, t)
}
func (t *HealthCheck) DeepCopy() *HealthCheck {
	if t == nil {
		return nil
	}
	out := new(HealthCheck)
	t.DeepCopyInto(out)
	return out
}
