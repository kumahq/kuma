package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	proto_util "github.com/kumahq/kuma/pkg/util/proto"
)

func (rl *RateLimit) SourceTags() (setList []SingleValueTagSet) {
	for _, selector := range rl.GetSources() {
		setList = append(setList, selector.Match)
	}
	return
}

func (m *RateLimit) UnmarshalJSON(data []byte) error {
	return proto_util.FromJSON(data, m)
}

func (m *RateLimit) MarshalJSON() ([]byte, error) {
	return proto_util.ToJSON(m)
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
