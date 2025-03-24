package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
)

// FallbackNodeHash is a hasher that will use as id either an item returned from a function or a fallback value
// this useful for having a fallback option when a nodeId is unknown or historically wasn't recorded
type FallBackNodeHash struct {
	GetIds    func() []string
	DefaultId string
}

var _ cache.NodeHash = &FallBackNodeHash{}

func (h *FallBackNodeHash) ID(node *envoy_core.Node) string {
	for _, id := range h.GetIds() {
		if id == node.Id {
			return id
		}
	}
	return h.DefaultId
}
