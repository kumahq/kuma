package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/golang/protobuf/ptypes/struct"

	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
)

type TagsMetadataConfigurer struct {
	Tags map[string]string
}

func (c *TagsMetadataConfigurer) Configure(l *envoy_api.Listener) error {
	if l.Metadata == nil {
		l.Metadata = &envoy_core.Metadata{}
	}
	if l.Metadata.FilterMetadata == nil {
		l.Metadata.FilterMetadata = map[string]*structpb.Struct{}
	}

	l.Metadata.FilterMetadata[envoy_metadata.TagsKey] = &structpb.Struct{
		Fields: envoy_metadata.MetadataFields(c.Tags),
	}
	return nil
}
