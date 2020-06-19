package tls

import (
	"fmt"
	"regexp"
	"strings"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/xds/envoy"
)

func SNIFromTags(tags envoy.Tags) string {
	nonServiceTags := tags.WithoutTag(mesh_proto.ServiceTag)
	var pairs []string
	for _, key := range nonServiceTags.Keys() {
		value := nonServiceTags[key]
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
	}
	service := tags[mesh_proto.ServiceTag]
	if len(pairs) == 0 {
		return service
	}
	return fmt.Sprintf("%s{%s}", service, strings.Join(pairs, ","))
}

func TagsFromSNI(sni string) map[string]string {
	r := regexp.MustCompile(`(.*)\{(.*)\}`)
	matches := r.FindStringSubmatch(sni)
	if len(matches) == 0 {
		return map[string]string{
			mesh_proto.ServiceTag: sni,
		}
	}
	service, tags := matches[1], matches[2]
	pairs := strings.Split(tags, ",")
	rv := map[string]string{
		mesh_proto.ServiceTag: service,
	}
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		rv[kv[0]] = kv[1]
	}
	return rv
}
