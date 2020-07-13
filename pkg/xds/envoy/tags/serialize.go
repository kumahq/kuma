package tags

import (
	"fmt"
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

func Serialize(tags mesh_proto.MultiValueTagSet) string {
	var pairs []string
	for _, key := range tags.Keys() {
		pairs = append(pairs, fmt.Sprintf("&%s=%s&", key, strings.Join(tags.Values(key), ",")))
	}
	return strings.Join(pairs, "")
}
