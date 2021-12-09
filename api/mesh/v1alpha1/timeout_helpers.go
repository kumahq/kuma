package v1alpha1

import (
	"time"

	"github.com/golang/protobuf/proto"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (x *Timeout_Conf) GetConnectTimeoutOrDefault(defaultConnectTimeout time.Duration) time.Duration {
	if x == nil {
		return defaultConnectTimeout
	}
	connectTimeout := x.GetConnectTimeout()
	if connectTimeout == nil {
		return defaultConnectTimeout
	}
	return connectTimeout.AsDuration()
}

func (m *Timeout) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *Timeout) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *Timeout) DeepCopyInto(out *Timeout) {
	proto.Merge(out, t)
}
func (t *Timeout) DeepCopy() *Timeout {
	if t == nil {
		return nil
	}
	out := new(Timeout)
	t.DeepCopyInto(out)
	return out
}
