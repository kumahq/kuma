package clusters

import (
	"net/url"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/config"
)

var _ config.Config = &ClustersConfig{}

type EndpointConfig struct {
	Address string `yaml:"address,omitempty"`
}

type ClusterConfig struct {
	Local   EndpointConfig `yaml:"local,omitempty"`
	Ingress EndpointConfig `yaml:"ingress,omitempty"`
}

// Clusters configuration
type ClustersConfig struct {
	Clusters []*ClusterConfig `yaml:"clusters,omitempty"`
}

func (g *ClustersConfig) Sanitize() {
}

func (g *ClustersConfig) Validate() error {
	for _, cluster := range g.Clusters {
		_, err := url.ParseRequestURI(cluster.Local.Address)
		if err != nil {
			return errors.Wrapf(err, "Invalid local url for cluster %s", cluster.Local.Address)
		}
		_, err = url.ParseRequestURI(cluster.Ingress.Address)
		if err != nil {
			return errors.Wrapf(err, "Invalid ingress url for cluster %s", cluster.Ingress.Address)
		}
	}
	return nil
}

func DefaultClustersConfig() *ClustersConfig {
	return &ClustersConfig{
		Clusters: []*ClusterConfig{},
	}
}
