package envoy

import (
	"fmt"
	"sort"
	"strings"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type Cluster struct {
	service           string
	name              string
	weight            uint32
	tags              Tags
	mesh              string
	isExternalService bool
	lb                *mesh_proto.TrafficRoute_LoadBalancer
	timeout           *mesh_proto.Timeout_Conf
}

func (c *Cluster) Service() string { return c.service }
func (c *Cluster) Name() string    { return c.name }
func (c *Cluster) Weight() uint32  { return c.weight }
func (c *Cluster) Tags() Tags      { return c.tags }

// Mesh returns a non-empty string only if the cluster is in a different mesh
// from the context.
func (c *Cluster) Mesh() string                              { return c.mesh }
func (c *Cluster) IsExternalService() bool                   { return c.isExternalService }
func (c *Cluster) LB() *mesh_proto.TrafficRoute_LoadBalancer { return c.lb }
func (c *Cluster) Timeout() *mesh_proto.Timeout_Conf         { return c.timeout }
func (c *Cluster) Hash() string                              { return fmt.Sprintf("%s-%s", c.name, c.tags.String()) }

func (c *Cluster) SetName(name string) {
	c.name = name
}

func (c *Cluster) SetMesh(mesh string) {
	c.mesh = mesh
}

type NewClusterOpt interface {
	apply(cluster *Cluster)
}

type newClusterOptFunc func(cluster *Cluster)

func (f newClusterOptFunc) apply(cluster *Cluster) {
	f(cluster)
}

func NewCluster(opts ...NewClusterOpt) Cluster {
	c := Cluster{}
	for _, opt := range opts {
		opt.apply(&c)
	}
	if err := c.validate(); err != nil {
		panic(err)
	}
	return c
}

func (c *Cluster) validate() error {
	if c.service == "" || c.name == "" {
		return errors.New("either WithService() or WithName() should be called")
	}
	return nil
}

func WithService(service string) NewClusterOpt {
	return newClusterOptFunc(func(cluster *Cluster) {
		cluster.service = service
		if len(cluster.name) == 0 {
			cluster.name = service
		}
	})
}

func WithName(name string) NewClusterOpt {
	return newClusterOptFunc(func(cluster *Cluster) {
		cluster.name = name
		if len(cluster.service) == 0 {
			cluster.service = name
		}
	})
}

func WithWeight(weight uint32) NewClusterOpt {
	return newClusterOptFunc(func(cluster *Cluster) {
		cluster.weight = weight
	})
}

func WithTags(tags Tags) NewClusterOpt {
	return newClusterOptFunc(func(cluster *Cluster) {
		cluster.tags = tags
	})
}

func WithTimeout(timeout *mesh_proto.Timeout_Conf) NewClusterOpt {
	return newClusterOptFunc(func(cluster *Cluster) {
		cluster.timeout = timeout
	})
}

func WithLB(lb *mesh_proto.TrafficRoute_LoadBalancer) NewClusterOpt {
	return newClusterOptFunc(func(cluster *Cluster) {
		cluster.lb = lb
	})
}

func WithExternalService(isExternalService bool) NewClusterOpt {
	return newClusterOptFunc(func(cluster *Cluster) {
		cluster.isExternalService = isExternalService
	})
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

type Service struct {
	name               string
	clusters           []Cluster
	hasExternalService bool
	tlsReady           bool
}

func (c *Service) Add(cluster Cluster) {
	c.clusters = append(c.clusters, cluster)
	if cluster.IsExternalService() {
		c.hasExternalService = true
	}
}

func (c *Service) Tags() []Tags {
	var result []Tags
	for _, cluster := range c.clusters {
		result = append(result, cluster.Tags())
	}
	return result
}

func (c *Service) HasExternalService() bool {
	return c.hasExternalService
}

func (c *Service) Clusters() []Cluster {
	return c.clusters
}

func (c *Service) TLSReady() bool {
	return c.tlsReady
}

type Services map[string]*Service

func (c Services) Sorted() []string {
	var keys []string
	for key := range c {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

type ServicesAccumulator struct {
	tlsReadiness map[string]bool
	services     map[string]*Service
}

func NewServicesAccumulator(tlsReadiness map[string]bool) ServicesAccumulator {
	return ServicesAccumulator{
		tlsReadiness: tlsReadiness,
		services:     map[string]*Service{},
	}
}

func (sa ServicesAccumulator) Services() Services {
	return sa.services
}

func (sa ServicesAccumulator) Add(clusters ...Cluster) {
	for _, c := range clusters {
		if sa.services[c.Service()] == nil {
			sa.services[c.Service()] = &Service{
				tlsReady: sa.tlsReadiness[c.Service()],
				name:     c.Service(),
			}
		}
		sa.services[c.Service()].Add(c)
	}
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
