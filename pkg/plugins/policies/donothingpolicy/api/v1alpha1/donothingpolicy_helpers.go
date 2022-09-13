package v1alpha1

import (
	"google.golang.org/protobuf/proto"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

func (x *DoNothingPolicy_From) GetDefaultAsProto() proto.Message {
	return x.Default
}

func (x *DoNothingPolicy) GetFromList() []core_xds.PolicyItem {
	var result []core_xds.PolicyItem
	for _, item := range x.From {
		result = append(result, item)
	}
	return result
}

func (x *DoNothingPolicy_To) GetDefaultAsProto() proto.Message {
	return x.Default
}

func (x *DoNothingPolicy) GetToList() []core_xds.PolicyItem {
	var result []core_xds.PolicyItem
	for _, item := range x.To {
		result = append(result, item)
	}
	return result
}
