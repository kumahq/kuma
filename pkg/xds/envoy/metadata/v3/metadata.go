package envoy

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

func EndpointMetadata(tags tags.Tags) *envoy_core.Metadata {
	tags = tags.WithoutTags(mesh_proto.ServiceTag) // service name is already in cluster name, we don't need it in metadata
	if len(tags) == 0 {
		return nil
	}
	fields := MetadataFields(tags)
	return &envoy_core.Metadata{
		FilterMetadata: map[string]*structpb.Struct{
			"envoy.lb": {
				Fields: fields,
			},
			"envoy.transport_socket_match": {
				Fields: fields,
			},
		},
	}
}

func LbMetadata(tags tags.Tags) *envoy_core.Metadata {
	tags = tags.WithoutTags(mesh_proto.ServiceTag) // service name is already in cluster name, we don't need it in metadata
	if len(tags) == 0 {
		return nil
	}
	fields := MetadataFields(tags)
	return &envoy_core.Metadata{
		FilterMetadata: map[string]*structpb.Struct{
			"envoy.lb": {
				Fields: fields,
			},
		},
	}
}

func MetadataFields(tags tags.Tags) map[string]*structpb.Value {
	fields := map[string]*structpb.Value{}
	for key, value := range tags {
		fields[key] = &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: value,
			},
		}
	}
	return fields
}

const (
	TagsKey   = "io.kuma.tags"
	LbTagsKey = "envoy.lb"
)

func ExtractTags(metadata *envoy_core.Metadata) tags.Tags {
	tags := tags.Tags{}
	for key, value := range metadata.GetFilterMetadata()[TagsKey].GetFields() {
		tags[key] = value.GetStringValue()
	}
	return tags
}

func ExtractLbTags(metadata *envoy_core.Metadata) tags.Tags {
	tags := tags.Tags{}
	for key, value := range metadata.GetFilterMetadata()[LbTagsKey].GetFields() {
		tags[key] = value.GetStringValue()
	}
	return tags
}
