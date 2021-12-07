package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *CircuitBreaker) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *CircuitBreaker) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *CircuitBreaker) DeepCopyInto(out *CircuitBreaker) {
	proto.Merge(out, t)
}
func (t *CircuitBreaker) DeepCopy() *CircuitBreaker {
	if t == nil {
		return nil
	}
	out := new(CircuitBreaker)
	t.DeepCopyInto(out)
	return out
}
