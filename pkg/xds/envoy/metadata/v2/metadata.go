package envoy

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	pstruct "github.com/golang/protobuf/ptypes/struct"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

func EndpointMetadata(tags envoy_common.Tags) *envoy_core.Metadata {
	tags = tags.WithoutTag(mesh_proto.ServiceTag) // service name is already in cluster name, we don't need it in metadata
	if len(tags) == 0 {
		return nil
	}
	fields := MetadataFields(tags)
	metadata := &envoy_core.Metadata{
		FilterMetadata: map[string]*pstruct.Struct{
			"envoy.lb": {
				Fields: fields,
			},
			"envoy.transport_socket_match": {
				Fields: fields,
			},
		},
	}
	return metadata
}

func LbMetadata(tags envoy_common.Tags) *envoy_core.Metadata {
	tags = tags.WithoutTag(mesh_proto.ServiceTag) // service name is already in cluster name, we don't need it in metadata
	if len(tags) == 0 {
		return nil
	}
	fields := MetadataFields(tags)
	metadata := &envoy_core.Metadata{
		FilterMetadata: map[string]*pstruct.Struct{
			"envoy.lb": {
				Fields: fields,
			},
		},
	}
	return metadata
}

func MetadataFields(tags envoy_common.Tags) map[string]*pstruct.Value {
	fields := map[string]*pstruct.Value{}
	for key, value := range tags {
		if key == mesh_proto.ServiceTag {
			continue
		}
		fields[key] = &pstruct.Value{
			Kind: &pstruct.Value_StringValue{
				StringValue: value,
			},
		}
	}
	return fields
}
