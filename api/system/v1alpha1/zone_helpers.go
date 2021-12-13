package v1alpha1

import (
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *Zone) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *Zone) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *Zone) DeepCopyInto(out *Zone) {
	util_proto.Merge(out, t)
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
