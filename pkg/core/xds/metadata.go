package xds

import "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

type DataplaneMetadata struct {
	AccessLogPort uint32
}

func DataplaneMetadataFromNode(node *core.Node) *DataplaneMetadata {
	metadata := DataplaneMetadata{}
	if portField := node.Metadata.Fields["accessLogPort"]; portField != nil {
		metadata.AccessLogPort = uint32(portField.GetNumberValue())
	}
	return &metadata
}
