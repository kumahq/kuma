package tls

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

func SNIFromTags(tags envoy_tags.Tags) string {
	extraTags := tags.WithoutTags(mesh_proto.ServiceTag).String()
	service := tags[mesh_proto.ServiceTag]
	if extraTags == "" {
		return service
	}
	return fmt.Sprintf("%s{%s}", service, extraTags)
}

func TagsFromSNI(sni string) (envoy_tags.Tags, error) {
	parts := strings.Split(sni, "{")
	if len(parts) > 2 {
		return nil, errors.New(fmt.Sprintf("cannot parse tags from sni: %s", sni))
	}
	if len(parts) == 1 {
		return envoy_tags.Tags{mesh_proto.ServiceTag: parts[0]}, nil
	}
	cleanedTags := strings.ReplaceAll(parts[1], "}", "")
	tags, err := envoy_tags.TagsFromString(cleanedTags)
	if err != nil {
		return nil, err
	}
	tags[mesh_proto.ServiceTag] = parts[0]
	return tags, nil
}
