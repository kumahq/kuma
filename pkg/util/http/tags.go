package http

import (
	"fmt"
	"strings"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
)

func SerializeTags(tags v1alpha1.MultiValueTagSet) string {
	var pairs []string
	for _, key := range tags.Keys() {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, strings.Join(tags.Values(key), ",")))
	}
	if len(pairs) == 0 {
		return ""
	}
	return "&" + strings.Join(pairs, "&") + "&"
}
