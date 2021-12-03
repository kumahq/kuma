package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	proto_util "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *CircuitBreaker) UnmarshalJSON(data []byte) error {
	return proto_util.FromJSON(data, m)
}

func (m *CircuitBreaker) MarshalJSON() ([]byte, error) {
	return proto_util.ToJSON(m)
}
func (t *CircuitBreaker) DeepCopyInto(out *CircuitBreaker) {
	proto.Merge(out, t)
}
