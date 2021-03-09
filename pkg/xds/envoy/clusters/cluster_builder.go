package clusters

import (
	"errors"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"

	"github.com/kumahq/kuma/pkg/xds/envoy"
	v2 "github.com/kumahq/kuma/pkg/xds/envoy/clusters/v2"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/clusters/v3"
)

// ClusterBuilderOpt is a configuration option for ClusterBuilder.
//
// The goal of ClusterBuilderOpt is to facilitate fluent ClusterBuilder API.
type ClusterBuilderOpt interface {
	// ApplyTo adds ClusterConfigurer(s) to the ClusterBuilder.
	ApplyTo(config *ClusterBuilderConfig)
}

func NewClusterBuilder(apiVersion envoy.APIVersion) *ClusterBuilder {
	return &ClusterBuilder{
		apiVersion: apiVersion,
	}
}

// ClusterBuilder is responsible for generating an Envoy cluster
// by applying a series of ClusterConfigurers.
type ClusterBuilder struct {
	apiVersion envoy.APIVersion
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
	case envoy.APIV2:
		cluster := envoy_api_v2.Cluster{}
		for _, configurer := range b.config.ConfigurersV2 {
			if err := configurer.Configure(&cluster); err != nil {
				return nil, err
			}
		}
		return &cluster, nil
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

// ClusterBuilderConfig holds configuration of a ClusterBuilder.
type ClusterBuilderConfig struct {
	// A series of ClusterConfigurers to apply to Envoy cluster.
	ConfigurersV2 []v2.ClusterConfigurer
	ConfigurersV3 []v3.ClusterConfigurer
}

// Add appends a given ClusterConfigurer to the end of the chain.
func (c *ClusterBuilderConfig) AddV2(configurer v2.ClusterConfigurer) {
	c.ConfigurersV2 = append(c.ConfigurersV2, configurer)
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
