package v1alpha1

import (
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *FaultInjection) SourceTags() (setList []SingleValueTagSet) {
	for _, selector := range m.GetSources() {
		setList = append(setList, selector.Match)
	}
	return
}

func (m *FaultInjection) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *FaultInjection) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *FaultInjection) DeepCopyInto(out *FaultInjection) {
	util_proto.Merge(out, t)
}
func (t *FaultInjection) DeepCopy() *FaultInjection {
	if t == nil {
		return nil
	}
	out := new(FaultInjection)
	t.DeepCopyInto(out)
	return out
}
