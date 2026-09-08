package tags

import (
	"fmt"
	"maps"
	"regexp"
	"sort"
	"strings"
)

const TagsHeaderName = "x-kuma-tags"

// Regex for parsing a split cluster name, see usages.
var splitClusterRegex = regexp.MustCompile("(.*)-[[:xdigit:]]{16}$")

type Tags map[string]string

func ServiceFromClusterName(name string) string {
	matchedGroups := splitClusterRegex.FindStringSubmatch(name)
	if len(matchedGroups) == 0 {
		return name
	}
	return matchedGroups[1]
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
