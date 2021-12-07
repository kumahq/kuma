package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	proto_util "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *FaultInjection) SourceTags() (setList []SingleValueTagSet) {
	for _, selector := range m.GetSources() {
		setList = append(setList, selector.Match)
	}
	return
}

func (m *FaultInjection) UnmarshalJSON(data []byte) error {
	return proto_util.FromJSON(data, m)
}

func (m *FaultInjection) MarshalJSON() ([]byte, error) {
	return proto_util.ToJSON(m)
}
func (t *FaultInjection) DeepCopyInto(out *FaultInjection) {
	proto.Merge(out, t)
}
func (t *FaultInjection) DeepCopy() *FaultInjection {
	if t == nil {
		return nil
	}
	out := new(FaultInjection)
	t.DeepCopyInto(out)
	return out
}
