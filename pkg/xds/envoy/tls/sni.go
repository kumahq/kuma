package tls

import (
	"fmt"
	"regexp"
	"strings"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/xds/envoy"
)

const sniRegexp = `(.*)\{(.*)\}`

func SNIFromTags(tags envoy.Tags) string {
	extraTags := tags.WithoutTag(mesh_proto.ServiceTag).String()
	service := tags[mesh_proto.ServiceTag]
	if extraTags == "" {
		return service
	}
	return fmt.Sprintf("%s{%s}", service, extraTags)
}

func TagsFromSNI(sni string) map[string]string {
	r := regexp.MustCompile(sniRegexp)
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
