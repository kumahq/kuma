package clusters

import (
	"net"
	"net/url"
	"strconv"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/config"
)

var _ config.Config = &ClustersConfig{}

type EndpointConfig struct {
	Address string `yaml:"address,omitempty"`
}

type ClusterConfig struct {
	Remote  EndpointConfig `yaml:"remote,omitempty"`
	Ingress EndpointConfig `yaml:"ingress,omitempty"`
}

type LBConfig struct {
	Address string `yaml:"address,omitempty"`
}

// Clusters configuration
type ClustersConfig struct {
	Clusters []*ClusterConfig `yaml:"clusters,omitempty"`
	LBConfig LBConfig         `yaml:"lbconfig,omitempty"`
}

func (g *ClustersConfig) Sanitize() {
}

func (g *ClustersConfig) Validate() error {
	for _, cluster := range g.Clusters {
		_, err := url.ParseRequestURI(cluster.Remote.Address)
		if err != nil {
			return errors.Wrapf(err, "Invalid remote url for cluster %s", cluster.Remote.Address)
		}
		_, port, err := net.SplitHostPort(cluster.Ingress.Address)
		if err != nil {
			return errors.Wrapf(err, "Invalid ingress address for cluster %s", cluster.Ingress.Address)
		}
		_, err = strconv.ParseUint(port, 10, 32)
		if err != nil {
			return errors.Wrapf(err, "Invalid ingress's port %s", port)
		}
	}
	return nil
}

func DefaultClustersConfig() *ClustersConfig {
	return &ClustersConfig{
		Clusters: []*ClusterConfig{},
	}
}
