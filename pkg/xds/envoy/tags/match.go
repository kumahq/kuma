package tags

import (
	"fmt"
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
