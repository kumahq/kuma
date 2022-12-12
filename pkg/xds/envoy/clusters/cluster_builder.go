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
	ApplyTo(config *ClusterBuilderConfig)
}

func NewClusterBuilder(apiVersion core_xds.APIVersion) *ClusterBuilder {
	return &ClusterBuilder{
		apiVersion: apiVersion,
	}
}

// ClusterBuilder is responsible for generating an Envoy cluster
// by applying a series of ClusterConfigurers.
type ClusterBuilder struct {
	apiVersion core_xds.APIVersion
	config     ClusterBuilderConfig
}

// Configure configures ClusterBuilder by adding individual ClusterConfigurers.
func (b *ClusterBuilder) Configure(opts ...ClusterBuilderOpt) *ClusterBuilder {
	for _, opt := range opts {
		opt.ApplyTo(&b.config)
	}
	return b
}

// Build generates an Envoy cluster by applying a series of ClusterConfigurers.
func (b *ClusterBuilder) Build() (envoy.NamedResource, error) {
	switch b.apiVersion {
	case envoy.APIV3:
		cluster := envoy_api.Cluster{}
		for _, configurer := range b.config.ConfigurersV3 {
			if err := configurer.Configure(&cluster); err != nil {
				return nil, err
			}
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

// ClusterBuilderConfig holds configuration of a ClusterBuilder.
type ClusterBuilderConfig struct {
	// A series of ClusterConfigurers to apply to Envoy cluster.
	ConfigurersV3 []v3.ClusterConfigurer
}

// Add appends a given ClusterConfigurer to the end of the chain.
func (c *ClusterBuilderConfig) AddV3(configurer v3.ClusterConfigurer) {
	c.ConfigurersV3 = append(c.ConfigurersV3, configurer)
}

// ClusterBuilderOptFunc is a convenience type adapter.
type ClusterBuilderOptFunc func(config *ClusterBuilderConfig)

func (f ClusterBuilderOptFunc) ApplyTo(config *ClusterBuilderConfig) {
	f(config)
}
