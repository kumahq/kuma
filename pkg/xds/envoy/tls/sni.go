package tls

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

func SNIFromTags(tags envoy.Tags) string {
	return SNIFromServiceAndTags(tags[mesh_proto.ServiceTag], tags.WithoutTag(mesh_proto.ServiceTag))
}

func SNIFromServiceAndTags(service string, tags envoy.Tags) string {
	tagsStr := tags.String()
	if tagsStr == "" {
		return service
	}
	return fmt.Sprintf("%s{%s}", service, tagsStr)
}
