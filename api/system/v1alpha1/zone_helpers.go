package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	proto_util "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *Zone) UnmarshalJSON(data []byte) error {
	return proto_util.FromJSON(data, m)
}

func (m *Zone) MarshalJSON() ([]byte, error) {
	return proto_util.ToJSON(m)
}
func (t *Zone) DeepCopyInto(out *Zone) {
	proto.Merge(out, t)
}
func (t *Zone) DeepCopy() *Zone {
	if t == nil {
		return nil
	}
	out := new(Zone)
	t.DeepCopyInto(out)
	return out
}

func (x *Zone) IsEnabled() bool {
	if x.Enabled == nil {
		return true
	}
	return x.Enabled.GetValue()
}
