package runtime

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
)

// Runtime represents initialized application state.
type Runtime interface {
	RuntimeInfo
	RuntimeContext
	ComponentManager
}

type RuntimeInfo interface {
	InstanceId() string
}

type RuntimeContext interface {
	Config() config.Config
	ResourceStore() core_store.ResourceStore
	DiscoverySources() []core_discovery.DiscoverySource
	XDS() core_xds.XdsContext
}

var _ Runtime = &runtime{}

type runtime struct {
	RuntimeInfo
	RuntimeContext
	ComponentManager
}

var _ RuntimeInfo = &runtimeInfo{}

type runtimeInfo struct {
	instanceId string
}

func (i *runtimeInfo) InstanceId() string {
	return i.instanceId
}

var _ RuntimeContext = &runtimeContext{}

type runtimeContext struct {
	cfg config.Config
	rs  core_store.ResourceStore
	dss []core_discovery.DiscoverySource
	xds core_xds.XdsContext
}

func (rc *runtimeContext) Config() config.Config {
	return rc.cfg
}
func (rc *runtimeContext) ResourceStore() core_store.ResourceStore {
	return rc.rs
}
func (rc *runtimeContext) DiscoverySources() []core_discovery.DiscoverySource {
	return rc.dss
}
func (rc *runtimeContext) XDS() core_xds.XdsContext {
	return rc.xds
}
