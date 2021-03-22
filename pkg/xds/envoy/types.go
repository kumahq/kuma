package envoy

import (
	"fmt"
	"sort"
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/pkg/errors"
)

type ClusterSubset struct {
	ClusterName       string
	Weight            uint32
	Tags              Tags
	IsExternalService bool
	Lb                *mesh_proto.TrafficRoute_LoadBalancer
	Timeout           *mesh_proto.Timeout_Conf
}

type Tags map[string]string
type TagsSlice []Tags
type TagKeys []string
type TagKeysSlice []TagKeys

func (t TagsSlice) ToTagKeysSlice() TagKeysSlice {
	out := []TagKeys{}
	for _, v := range t {
		out = append(out, v.Keys())
	}
	return out
}

// Transform applies each transformer to each TagKeys and returns a sorted unique TagKeysSlice.
func (t TagKeysSlice) Transform(transformers ...TagKeyTransformer) TagKeysSlice {
	allSlices := map[string]TagKeys{}
	for _, tagKeys := range t {
		res := tagKeys.Transform(transformers...)
		if len(res) > 0 {
			h := strings.Join(res, ", ")
			allSlices[h] = res
		}
	}
	out := TagKeysSlice{}
	for _, n := range allSlices {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool {
		for k := 0; k < len(out[i]) && k < len(out[j]); k++ {
			if out[i][k] != out[j][k] {
				return out[i][k] < out[j][k]
			}
		}
		return len(out[i]) < len(out[j])
	})
	return out
}

type TagKeyTransformer interface {
	Apply(slice TagKeys) TagKeys
}
type TagKeyTransformerFunc func(slice TagKeys) TagKeys

func (f TagKeyTransformerFunc) Apply(slice TagKeys) TagKeys {
	return f(slice)
}

// Transform applies a list of transformers on the tag keys and return a new set of keys (always return sorted, unique sets).
func (t TagKeys) Transform(transformers ...TagKeyTransformer) TagKeys {
	tmp := t
	for _, tr := range transformers {
		tmp = tr.Apply(tmp)
	}
	// Make tags unique and sorted
	tagSet := map[string]bool{}
	out := TagKeys{}
	for _, n := range tmp {
		if !tagSet[n] {
			tagSet[n] = true
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}

func Without(tags ...string) TagKeyTransformer {
	tagSet := map[string]bool{}
	for _, t := range tags {
		tagSet[t] = true
	}
	return TagKeyTransformerFunc(func(slice TagKeys) TagKeys {
		out := []string{}
		for _, t := range slice {
			if !tagSet[t] {
				out = append(out, t)
			}
		}
		return out
	})
}

func With(tags ...string) TagKeyTransformer {
	return TagKeyTransformerFunc(func(slice TagKeys) TagKeys {
		res := make([]string, len(tags)+len(slice))
		copy(res, slice)
		copy(res[len(slice):], tags)
		return res
	})
}

func (t Tags) WithoutTags(tags ...string) Tags {
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
	for tagName, tagValue := range t {
		result[tagName] = tagValue
	}
	for i := 0; i < len(keysAndValues); {
		key, value := keysAndValues[i], keysAndValues[i+1]
		result[key] = value
		i += 2
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

func TagsFromString(tagsString string) (Tags, error) {
	result := Tags{}
	tagPairs := strings.Split(tagsString, ",")
	for _, pair := range tagPairs {
		split := strings.Split(pair, "=")
		if len(split) != 2 {
			return nil, errors.New("invalid format of tags, pairs should be separated by , and key should be separated from value by =")
		}
		result[split[0]] = split[1]
	}
	return result, nil
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

func TagKeySlice(tags []Tags) TagKeysSlice {
	r := make([]TagKeys, len(tags))
	for i := range tags {
		r[i] = tags[i].Keys()
	}
	return r
}

type Cluster struct {
	subsets            []ClusterSubset
	hasExternalService bool
	lb                 *mesh_proto.TrafficRoute_LoadBalancer
	timeout            *mesh_proto.Timeout_Conf
}

func (c *Cluster) Add(subset ClusterSubset) {
	c.subsets = append(c.subsets, subset)
	if subset.IsExternalService {
		c.hasExternalService = true
	}
	c.lb = subset.Lb
	c.timeout = subset.Timeout
}

func (c *Cluster) Tags() []Tags {
	var result []Tags
	for _, info := range c.subsets {
		result = append(result, info.Tags)
	}
	return result
}

func (c *Cluster) HasExternalService() bool {
	return c.hasExternalService
}

func (c *Cluster) Subsets() []ClusterSubset {
	return c.subsets
}

func (c *Cluster) Timeout() *mesh_proto.Timeout_Conf {
	return c.timeout
}

type Clusters map[string]*Cluster

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
		if c[info.ClusterName] == nil {
			c[info.ClusterName] = &Cluster{}
		}
		c[info.ClusterName].Add(info)
	}
}

func (c Clusters) Get(name string) *Cluster {
	return c[name]
}

func (c Clusters) Tags(name string) []Tags {
	return c[name].Tags()
}

func (c Clusters) Lb(name string) *mesh_proto.TrafficRoute_LoadBalancer {
	return c[name].lb
}

type NamedResource interface {
	envoy_types.Resource
	GetName() string
}

type TrafficDirection string

const (
	TrafficDirectionOutbound    TrafficDirection = "OUTBOUND"
	TrafficDirectionInbound     TrafficDirection = "INBOUND"
	TrafficDirectionUnspecified TrafficDirection = "UNSPECIFIED"
)

type StaticEndpointPath struct {
	Path             string
	ClusterName      string
	RewritePath      string
	Header           string
	HeaderExactMatch string
}
