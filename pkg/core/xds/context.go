package xds

import (
	"fmt"

	"github.com/Kong/kuma/pkg/core"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	envoy_log "github.com/envoyproxy/go-control-plane/pkg/log"
	"github.com/go-logr/logr"
)

type XdsContext interface {
	Hasher() envoy_cache.NodeHash
	Cache() envoy_cache.SnapshotCache
}

func NewXdsContext() XdsContext {
	return newXdsContext("xds-server", true)
}

func newXdsContext(name string, ads bool) XdsContext {
	log := core.Log.WithName(name)
	hasher := hasher{log}
	logger := logger{log}
	cache := envoy_cache.NewSnapshotCache(ads, hasher, logger)
	return &xdsContext{
		NodeHash:      hasher,
		Logger:        logger,
		SnapshotCache: cache,
	}
}

var _ XdsContext = &xdsContext{}

type xdsContext struct {
	envoy_cache.NodeHash
	envoy_log.Logger
	envoy_cache.SnapshotCache
}

func (c *xdsContext) Hasher() envoy_cache.NodeHash {
	return c.NodeHash
}

func (c *xdsContext) Cache() envoy_cache.SnapshotCache {
	return c.SnapshotCache
}

var _ envoy_cache.NodeHash = &hasher{}

type hasher struct {
	log logr.Logger
}

func (h hasher) ID(node *envoy_core.Node) string {
	if node == nil {
		return "unknown"
	}
	proxyId, err := ParseProxyId(node)
	if err != nil {
		h.log.Error(err, "failed to parse Proxy ID", "node", node)
		return "unknown"
	}
	return proxyId.String()
}

var _ envoy_log.Logger = &logger{}

type logger struct {
	log logr.Logger
}

func (l logger) Infof(format string, args ...interface{}) {
	l.log.V(1).Info(fmt.Sprintf(format, args...))
}
func (l logger) Errorf(format string, args ...interface{}) {
	l.log.Error(fmt.Errorf(format, args...), "")
}
