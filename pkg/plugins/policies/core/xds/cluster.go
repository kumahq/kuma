package xds

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

type Cluster struct {
	service           string
	name              string
	tags              tags.Tags
	mesh              string
	isExternalService bool
	statName          string
}

func (c *Cluster) Service() string { return c.service }
func (c *Cluster) Name() string    { return c.name }
func (c *Cluster) Tags() tags.Tags { return c.tags }

// Mesh returns a non-empty string only if the cluster is in a different mesh
// from the context.
func (c *Cluster) Mesh() string            { return c.mesh }
func (c *Cluster) IsExternalService() bool { return c.isExternalService }
func (c *Cluster) Hash() string            { return fmt.Sprintf("%s-%s", c.name, c.tags.String()) }
func (c *Cluster) StatName() string        { return c.statName }

type NewClusterOpt interface {
	apply(cluster *Cluster)
}

type newClusterOptFunc func(cluster *Cluster)

func (f newClusterOptFunc) apply(cluster *Cluster) {
	f(cluster)
}

type ClusterBuilder struct {
	opts []NewClusterOpt
}

func NewClusterBuilder() *ClusterBuilder {
	return &ClusterBuilder{}
}

func (b *ClusterBuilder) Build() *Cluster {
	c := &Cluster{}
	for _, opt := range b.opts {
		opt.apply(c)
	}
	if err := c.validate(); err != nil {
		panic(err)
	}
	return c
}

func (b *ClusterBuilder) WithService(service string) *ClusterBuilder {
	b.opts = append(b.opts, newClusterOptFunc(func(cluster *Cluster) {
		cluster.service = service
		if len(cluster.name) == 0 {
			cluster.name = service
		}
	}))
	return b
}

func (b *ClusterBuilder) WithName(name string) *ClusterBuilder {
	b.opts = append(b.opts, newClusterOptFunc(func(cluster *Cluster) {
		cluster.name = name
		if len(cluster.service) == 0 {
			cluster.service = name
		}
	}))
	return b
}

func (b *ClusterBuilder) WithMesh(mesh string) *ClusterBuilder {
	b.opts = append(b.opts, newClusterOptFunc(func(cluster *Cluster) {
		cluster.mesh = mesh
	}))
	return b
}

func (b *ClusterBuilder) WithTags(tags tags.Tags) *ClusterBuilder {
	b.opts = append(b.opts, newClusterOptFunc(func(cluster *Cluster) {
		cluster.tags = tags
	}))
	return b
}

func (b *ClusterBuilder) WithExternalService(isExternalService bool) *ClusterBuilder {
	b.opts = append(b.opts, newClusterOptFunc(func(cluster *Cluster) {
		cluster.isExternalService = isExternalService
	}))
	return b
}

func (b *ClusterBuilder) WithStatName(statName string) *ClusterBuilder {
	b.opts = append(b.opts, newClusterOptFunc(func(cluster *Cluster) {
		cluster.statName = statName
	}))
	return b
}

func (c *Cluster) validate() error {
	if c.service == "" || c.name == "" {
		return errors.New("either WithService() or WithName() should be called")
	}
	return nil
}
