package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

// ClusterConfigurer is responsible for configuring a single aspect of the entire Envoy cluster,
// such as filter chain, transparent proxying, etc.
type ClusterConfigurer interface {
	// Configure configures a single aspect on a given Envoy cluster.
	Configure(cluster *envoy_api.Cluster) error
}

// ClusterBuilderOpt is a configuration option for ClusterBuilder.
//
// The goal of ClusterBuilderOpt is to facilitate fluent ClusterBuilder API.
type ClusterBuilderOpt interface {
	// ApplyTo adds ClusterConfigurer(s) to the ClusterBuilder.
	ApplyTo(config *ClusterBuilderConfig)
}

func NewClusterBuilder() *ClusterBuilder {
	return &ClusterBuilder{}
}

// ClusterBuilder is responsible for generating an Envoy cluster
// by applying a series of ClusterConfigurers.
type ClusterBuilder struct {
	config ClusterBuilderConfig
}

// Configure configures ClusterBuilder by adding individual ClusterConfigurers.
func (b *ClusterBuilder) Configure(opts ...ClusterBuilderOpt) *ClusterBuilder {
	for _, opt := range opts {
		opt.ApplyTo(&b.config)
	}
	return b
}

// Build generates an Envoy cluster by applying a series of ClusterConfigurers.
func (b *ClusterBuilder) Build() (*envoy_api.Cluster, error) {
	cluster := envoy_api.Cluster{}
	for _, configurer := range b.config.Configurers {
		if err := configurer.Configure(&cluster); err != nil {
			return nil, err
		}
	}
	return &cluster, nil
}

// ClusterBuilderConfig holds configuration of a ClusterBuilder.
type ClusterBuilderConfig struct {
	// A series of ClusterConfigurers to apply to Envoy cluster.
	Configurers []ClusterConfigurer
}

// Add appends a given ClusterConfigurer to the end of the chain.
func (c *ClusterBuilderConfig) Add(configurer ClusterConfigurer) {
	c.Configurers = append(c.Configurers, configurer)
}

// ClusterBuilderOptFunc is a convenience type adapter.
type ClusterBuilderOptFunc func(config *ClusterBuilderConfig)

func (f ClusterBuilderOptFunc) ApplyTo(config *ClusterBuilderConfig) {
	f(config)
}
