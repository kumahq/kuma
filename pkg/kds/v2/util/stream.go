package util

import (
	"strconv"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/structpb"
)

const streamMetadataKey = "kds-stream-id"

func FillStreamMetadata(streamID int64, node *envoy_core.Node) {
	if node.Metadata == nil {
		node.Metadata = &structpb.Struct{}
	}
	if node.Metadata.Fields == nil {
		node.Metadata.Fields = map[string]*structpb.Value{}
	}
	node.Metadata.Fields[streamMetadataKey] = structpb.NewStringValue(strconv.FormatInt(streamID, 10))
}

func StreamFromMetadata(node *envoy_core.Node) (string, bool) {
	val, ok := node.GetMetadata().GetFields()[streamMetadataKey]
	return val.GetStringValue(), ok
}
