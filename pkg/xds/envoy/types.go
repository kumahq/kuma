package envoy

import (
	"context"
	"fmt"
	"sort"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

type Cluster struct {
	service           string
	name              string
	weight            uint32
	tags              tags.Tags
	mesh              string
	isExternalService bool
	lb                *mesh_proto.TrafficRoute_LoadBalancer
	timeout           *mesh_proto.Timeout_Conf
}

func (c *Cluster) Service() string { return c.service }
func (c *Cluster) Name() string    { return c.name }
func (c *Cluster) Weight() uint32  { return c.weight }
func (c *Cluster) Tags() tags.Tags { return c.tags }

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

func WithTags(tags tags.Tags) NewClusterOpt {
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

func (c *Service) Tags() []tags.Tags {
	var result []tags.Tags
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

type CLACache interface {
	GetCLA(ctx context.Context, meshName, meshHash string, cluster Cluster, apiVersion core_xds.APIVersion, endpointMap core_xds.EndpointMap) (proto.Message, error)
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
