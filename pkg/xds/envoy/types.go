package envoy

import (
	"fmt"
	"sort"
	"strings"
)

type ClusterSubset struct {
	ClusterName string
	Weight      uint32
	Tags        Tags
}

type Tags map[string]string

func (t Tags) WithoutTag(tag string) Tags {
	result := Tags{}
	for tagName, tagValue := range t {
		if tag != tagName {
			result[tagName] = tagValue
		}
	}
	return result
}

func (t Tags) Keys() []string {
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

func DistinctTags(tags []Tags) []Tags {
	used := map[string]bool{}
	var result []Tags
	for _, tag := range tags {
		str := tag.String()
		if !used[str] {
			result = append(result, tag)
			used[str] = true
		}
	}
	return result
}

type Clusters map[string][]ClusterSubset

func (c Clusters) ClusterNames() []string {
	var keys []string
	for key := range c {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (c Clusters) Add(infos ...ClusterSubset) {
	for _, info := range infos {
		c[info.ClusterName] = append(c[info.ClusterName], info)
	}
}

func (c Clusters) Tags(name string) []Tags {
	var result []Tags
	for _, info := range c[name] {
		result = append(result, info.Tags)
	}
	return result
}
