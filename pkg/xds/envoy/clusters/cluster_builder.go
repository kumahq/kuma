package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/pkg/errors"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/clusters/v3"
)

// ClusterBuilderOpt is a configuration option for ClusterBuilder.
//
// The goal of ClusterBuilderOpt is to facilitate fluent ClusterBuilder API.
type ClusterBuilderOpt interface {
	// ApplyTo adds ClusterConfigurer(s) to the ClusterBuilder.
	ApplyTo(builder *ClusterBuilder)
}

func NewClusterBuilder(apiVersion core_xds.APIVersion, name string) *ClusterBuilder {
	return &ClusterBuilder{
		apiVersion: apiVersion,
		name:       name,
	}
}

// ClusterBuilder is responsible for generating an Envoy cluster
// by applying a series of ClusterConfigurers.
type ClusterBuilder struct {
	apiVersion core_xds.APIVersion
	// A series of ClusterConfigurers to apply to Envoy cluster.
	configurers []v3.ClusterConfigurer
	name        string
}

// Configure configures ClusterBuilder by adding individual ClusterConfigurers.
func (b *ClusterBuilder) Configure(opts ...ClusterBuilderOpt) *ClusterBuilder {
	for _, opt := range opts {
		opt.ApplyTo(b)
	}
	return b
}

// Build generates an Envoy cluster by applying a series of ClusterConfigurers.
func (b *ClusterBuilder) Build() (envoy.NamedResource, error) {
	switch b.apiVersion {
	case envoy.APIV3:
		cluster := envoy_api.Cluster{
			Name: b.name,
		}
		for _, configurer := range b.configurers {
			if err := configurer.Configure(&cluster); err != nil {
				return nil, err
			}
		}
		if cluster.GetName() == "" {
			return nil, errors.New("cluster name is undefined")
		}
		return &cluster, nil
	default:
		return nil, errors.New("unknown API")
	}
}

func (b *ClusterBuilder) MustBuild() envoy.NamedResource {
	cluster, err := b.Build()
	if err != nil {
		panic(errors.Wrap(err, "failed to build Envoy Cluster").Error())
	}

	return cluster
}

// AddConfigurer appends a given ClusterConfigurer to the end of the chain.
func (b *ClusterBuilder) AddConfigurer(configurer v3.ClusterConfigurer) {
	b.configurers = append(b.configurers, configurer)
}

// WithName sets the name for the cluster
func (b *ClusterBuilder) WithName(name string) {
	b.name = name
}

// ClusterBuilderOptFunc is a convenience type adapter.
type ClusterBuilderOptFunc func(config *ClusterBuilder)

func (f ClusterBuilderOptFunc) ApplyTo(builder *ClusterBuilder) {
	f(builder)
}
