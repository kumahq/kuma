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

func MetadataListValues(tags []tags.Tags) *structpb.ListValue {
	list := &structpb.ListValue{}
	for _, tag := range tags {
		list.Values = append(list.Values, &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: &structpb.Struct{
					Fields: MetadataFields(tag),
				},
			},
		})
	}
	return list
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

const TagsKey = "io.kuma.tags"
const RouteTagsKey = "io.kuma.route.tags"

func ExtractTags(metadata *envoy_core.Metadata) tags.Tags {
	tags := tags.Tags{}
	for key, value := range metadata.GetFilterMetadata()[TagsKey].GetFields() {
		tags[key] = value.GetStringValue()
	}
	return tags
}

func ExtractListOfTags(metadata *envoy_core.Metadata) []tags.Tags {
	allTags := []tags.Tags{}
	for _, value := range metadata.GetFilterMetadata()[RouteTagsKey].GetFields() {
		val := value.GetListValue()
		for _, entry := range val.GetValues() {
			metadataTags := entry.GetStructValue()
			selectorTags := tags.Tags{}
			for header, headerValue := range metadataTags.GetFields() {
				selectorTags[header] = headerValue.GetStringValue()
			}
			allTags = append(allTags, selectorTags)
		}
	}
	return allTags
}
