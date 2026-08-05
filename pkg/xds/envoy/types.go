package envoy

import (
	"context"
	"slices"
	"sort"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"google.golang.org/protobuf/proto"

	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/tags"
)

type Cluster interface {
	Service() string
	Name() string
	SNI() string
	Tags() tags.Tags
	Hash() string
	IsExternalService() bool
}

type Split interface {
	ClusterName() string
	Weight() uint32
	LBMetadata() tags.Tags
	HasExternalService() bool
}

type Service struct {
	name               string
	clusters           []Cluster
	hasExternalService bool
	tlsReady           bool
	backendRef         *resolve.ResolvedBackendRef
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

func (c *Service) BackendRef() *resolve.ResolvedBackendRef {
	return c.backendRef
}

type Services map[string]*Service

func (s Services) Sorted() []string {
	var keys []string
	for key := range s {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (s Services) Clusters() []Cluster {
	var clusters []Cluster

	for _, serviceName := range s.Sorted() {
		clusters = slices.Concat(clusters, s[serviceName].Clusters())
	}

	return clusters
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

func (sa ServicesAccumulator) AddBackendRef(backendRef *resolve.ResolvedBackendRef, cluster Cluster) {
	if sa.services[cluster.Service()] == nil {
		sa.services[cluster.Service()] = &Service{
			tlsReady:   sa.tlsReadiness[cluster.Service()],
			name:       cluster.Service(),
			backendRef: backendRef,
		}
	}
	// prioritize backendRef pointing to real resource
	if backendRef.ReferencesRealResource() && !sa.services[cluster.Service()].backendRef.ReferencesRealResource() {
		sa.services[cluster.Service()].backendRef = backendRef
	}
	sa.services[cluster.Service()].Add(cluster)
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
