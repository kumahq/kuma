package runtime

import (
	"context"

	kuma_cp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/kuma-cp"
	builtin_ca "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/ca/builtin"
	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	core_manager "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/manager"
	secret_manager "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/manager"
	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
)

// Runtime represents initialized application state.
type Runtime interface {
	RuntimeInfo
	RuntimeContext
	ComponentManager
}

type RuntimeInfo interface {
	GetInstanceId() string
}

type RuntimeContext interface {
	Config() kuma_cp.Config
	DiscoverySources() []core_discovery.DiscoverySource
	XDS() core_xds.XdsContext
	ResourceManager() core_manager.ResourceManager
	SecretManager() secret_manager.SecretManager
	BuiltinCaManager() builtin_ca.BuiltinCaManager
	Extensions() context.Context
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

func (i *runtimeInfo) GetInstanceId() string {
	return i.instanceId
}

var _ RuntimeContext = &runtimeContext{}

type runtimeContext struct {
	cfg kuma_cp.Config
	rm  core_manager.ResourceManager
	sm  secret_manager.SecretManager
	bcm builtin_ca.BuiltinCaManager
	dss []core_discovery.DiscoverySource
	xds core_xds.XdsContext
	ext context.Context
}

func (rc *runtimeContext) Config() kuma_cp.Config {
	return rc.cfg
}
func (rc *runtimeContext) DiscoverySources() []core_discovery.DiscoverySource {
	return rc.dss
}
func (rc *runtimeContext) XDS() core_xds.XdsContext {
	return rc.xds
}
func (rc *runtimeContext) ResourceManager() core_manager.ResourceManager {
	return rc.rm
}
func (rc *runtimeContext) SecretManager() secret_manager.SecretManager {
	return rc.sm
}
func (rc *runtimeContext) BuiltinCaManager() builtin_ca.BuiltinCaManager {
	return rc.bcm
}
func (rc *runtimeContext) Extensions() context.Context {
	return rc.ext
}
