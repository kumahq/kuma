package util

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/structpb"
)

const tenantMetadataKey = "tenant"

func FillTenantMetadata(tenantID string, node *envoy_core.Node) {
	if node.Metadata == nil {
		node.Metadata = &structpb.Struct{}
	}
	if node.Metadata.Fields == nil {
		node.Metadata.Fields = map[string]*structpb.Value{}
	}
	node.Metadata.Fields[tenantMetadataKey] = structpb.NewStringValue(tenantID)
}

func TenantFromMetadata(node *envoy_core.Node) (string, bool) {
	val, ok := node.GetMetadata().GetFields()[tenantMetadataKey]
	return val.GetStringValue(), ok
}
