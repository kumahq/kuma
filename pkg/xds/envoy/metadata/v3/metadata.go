package envoy

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/xds/envoy/tags"
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

// EndpointMetadataWithLabels builds Envoy endpoint filter metadata that includes
// inbound tags (under the "envoy.lb" key) and, when inbound tags are absent,
// pod/workload labels (under the "io.kuma.labels" key). This allows
// AffinityTags to fall back to pod labels when inbound tags are absent due to
// KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED, without bloating metadata when tags
// are present.
func EndpointMetadataWithLabels(t tags.Tags, labels map[string]string) *envoy_core.Metadata {
	meta := EndpointMetadata(t)
	// Labels are used strictly as a fallback when inbound tags are unavailable.
	// If we already have tag-derived metadata or there are no labels, return as-is.
	if meta != nil || len(labels) == 0 {
		return meta
	}
	labelFields := make(map[string]*structpb.Value, len(labels))
	for k, v := range labels {
		labelFields[k] = structpb.NewStringValue(v)
	}
	return &envoy_core.Metadata{
		FilterMetadata: map[string]*structpb.Struct{
			LbLabelsKey: {
				Fields: labelFields,
			},
		},
	}
}

// ExtractLbLabels reads pod/workload labels from Envoy endpoint filter metadata.
// Labels are stored under the "io.kuma.labels" key (separate from inbound tags)
// and are available even when KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED is true.
func ExtractLbLabels(metadata *envoy_core.Metadata) tags.Tags {
	if metadata == nil {
		return nil
	}
	structVal, ok := metadata.GetFilterMetadata()[LbLabelsKey]
	if !ok || structVal == nil {
		return nil
	}
	result := tags.Tags{}
	for key, value := range structVal.GetFields() {
		result[key] = value.GetStringValue()
	}
	return result
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
	// LbLabelsKey is the filter metadata key under which pod/workload labels are
	// stored. Unlike LbTagsKey (inbound tags), labels remain available even when
	// KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED is set to true.
	LbLabelsKey = "io.kuma.labels"
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
