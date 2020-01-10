package reconcile

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"

	mads_cache "github.com/Kong/kuma/pkg/mads/cache"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

func NewSnapshotGenerator() SnapshotGenerator {
	return &snapshotGenerator{}
}

type snapshotGenerator struct {
}

func (_ snapshotGenerator) GenerateSnapshot(*envoy_core.Node) (util_xds.Snapshot, error) {
	return mads_cache.NewSnapshot("", map[string]envoy_cache.Resource{}), nil
}
