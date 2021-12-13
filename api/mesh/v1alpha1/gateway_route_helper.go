package v1alpha1

import (
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *GatewayRoute) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *GatewayRoute) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *GatewayRoute) DeepCopyInto(out *GatewayRoute) {
	util_proto.Merge(out, t)
}
func (t *GatewayRoute) DeepCopy() *GatewayRoute {
	if t == nil {
		return nil
	}
	out := new(GatewayRoute)
	t.DeepCopyInto(out)
	return out
}
