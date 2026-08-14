package tags

import (
	"crypto/sha256"
	"fmt"
	"maps"
	"regexp"
	"sort"
	"strings"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	util_maps "github.com/kumahq/kuma/v3/pkg/util/maps"
)

const TagsHeaderName = "x-kuma-tags"

// Format of split cluster name and regex for parsing it, see usages
const splitClusterFmtString = "%s-%x"

var splitClusterRegex = regexp.MustCompile("(.*)-[[:xdigit:]]{16}$")

type Tags map[string]string

func ServiceFromClusterName(name string) string {
	matchedGroups := splitClusterRegex.FindStringSubmatch(name)
	if len(matchedGroups) == 0 {
		return name
	}
	return matchedGroups[1]
}

// DestinationClusterName generates a unique cluster name for the
// destination. identifyingTags are useful for adding extra metadata outside of just tags. Tags must at least contain `kuma.io/service`
func (t Tags) DestinationClusterName(
	additionalIdentifyingTags map[string]string,
) (string, error) {
	serviceName := t[mesh_proto.ServiceTag]
	if serviceName == "" {
		return "", fmt.Errorf("missing %s tag", mesh_proto.ServiceTag)
	}

	// If there's no tags other than serviceName just return the serviceName
	if len(additionalIdentifyingTags) == 0 && len(t) == 1 {
		return serviceName, nil
	}

	// If cluster is splitting the target service with selector tags,
	// hash the tag names to generate a unique cluster name.
	h := sha256.New()

	for _, k := range util_maps.SortedKeys(t) {
		h.Write([]byte(k))
		h.Write([]byte(t[k]))
	}
	for _, k := range util_maps.SortedKeys(additionalIdentifyingTags) {
		h.Write([]byte(k))
		h.Write([]byte(additionalIdentifyingTags[k]))
	}

	// The qualifier is 16 hex digits. Unscientifically balancing the length
	// of the hex against the likelihood of collisions.
	// Note: policy configuration is sensitive to this format!
	return fmt.Sprintf(splitClusterFmtString, serviceName, h.Sum(nil)[:8]), nil
}

func (t Tags) WithoutTags(tags ...string) Tags {
	if t == nil {
		return nil
	}
	tagSet := map[string]bool{}
	for _, t := range tags {
		tagSet[t] = true
	}
	result := Tags{}
	for tagName, tagValue := range t {
		if !tagSet[tagName] {
			result[tagName] = tagValue
		}
	}
	return result
}

func (t Tags) WithTags(keysAndValues ...string) Tags {
	result := Tags{}
	maps.Copy(result, t)
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		key, value := keysAndValues[i], keysAndValues[i+1]
		result[key] = value
	}
	return result
}

func (t Tags) Keys() TagKeys {
	var keys []string
	for key := range t {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (t Tags) String() string {
	var pairs []string
	for _, key := range t.Keys() {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, t[key]))
	}
	return strings.Join(pairs, ",")
}

type TagKeys []string
