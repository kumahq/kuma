package v1alpha1

import (
	"github.com/golang/protobuf/proto"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (rl *RateLimit) SourceTags() (setList []SingleValueTagSet) {
	for _, selector := range rl.GetSources() {
		setList = append(setList, selector.Match)
	}
	return
}

func (m *RateLimit) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *RateLimit) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *RateLimit) DeepCopyInto(out *RateLimit) {
	proto.Merge(out, t)
}
func (t *RateLimit) DeepCopy() *RateLimit {
	if t == nil {
		return nil
	}
	out := new(RateLimit)
	t.DeepCopyInto(out)
	return out
}
