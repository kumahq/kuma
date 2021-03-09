package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_log "github.com/envoyproxy/go-control-plane/pkg/log"
	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/xds"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
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
	logger := util_xds.NewLogger(log)
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
	proxyId, err := xds.ParseProxyIdFromString(node.GetId())
	if err != nil {
		h.log.Error(err, "failed to parse Proxy ID", "node", node)
		return "unknown"
	}
	return proxyId.String()
}
