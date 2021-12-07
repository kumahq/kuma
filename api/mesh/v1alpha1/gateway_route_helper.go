package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	proto_util "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *GatewayRoute) UnmarshalJSON(data []byte) error {
	return proto_util.FromJSON(data, m)
}

func (m *GatewayRoute) MarshalJSON() ([]byte, error) {
	return proto_util.ToJSON(m)
}
func (t *GatewayRoute) DeepCopyInto(out *GatewayRoute) {
	proto.Merge(out, t)
}
func (t *GatewayRoute) DeepCopy() *GatewayRoute {
	if t == nil {
		return nil
	}
	out := new(GatewayRoute)
	t.DeepCopyInto(out)
	return out
}
