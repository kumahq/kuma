package envoy

import "sort"

type ClusterInfo struct {
	Name   string
	Weight uint32
	Tags   Tags
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
	for key, _ := range t {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
