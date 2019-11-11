package xds

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

type DataplaneMetadata struct {
	DataplaneTokenPath string
}

func DataplaneMetadataFromNode(node *envoy_core.Node) *DataplaneMetadata {
	metadata := DataplaneMetadata{}
	if node.Metadata == nil {
		return &metadata
	}
	if field := node.Metadata.Fields["dataplaneTokenPath"]; field != nil {
		metadata.DataplaneTokenPath = field.GetStringValue()
	}
	return &metadata
}
