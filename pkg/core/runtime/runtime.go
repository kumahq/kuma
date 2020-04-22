package runtime

import (
	"context"

	"github.com/Kong/kuma/pkg/core/ca"

	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	secret_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

// Runtime represents initialized application state.
type Runtime interface {
	RuntimeInfo
	RuntimeContext
	component.Manager
}

type RuntimeInfo interface {
	GetInstanceId() string
}

type RuntimeContext interface {
	Config() kuma_cp.Config
	XDS() core_xds.XdsContext
	ResourceManager() core_manager.ResourceManager
	ReadOnlyResourceManager() core_manager.ReadOnlyResourceManager
	SecretManager() secret_manager.SecretManager
	CaManagers() ca.CaManagers
	Extensions() context.Context
}

var _ Runtime = &runtime{}

type runtime struct {
	RuntimeInfo
	RuntimeContext
	component.Manager
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
	rom core_manager.ReadOnlyResourceManager
	sm  secret_manager.SecretManager
	cam ca.CaManagers
	xds core_xds.XdsContext
	ext context.Context
}

func (rc *runtimeContext) CaManagers() ca.CaManagers {
	return rc.cam
}
func (rc *runtimeContext) Config() kuma_cp.Config {
	return rc.cfg
}
func (rc *runtimeContext) XDS() core_xds.XdsContext {
	return rc.xds
}
func (rc *runtimeContext) ResourceManager() core_manager.ResourceManager {
	return rc.rm
}
func (rc *runtimeContext) ReadOnlyResourceManager() core_manager.ReadOnlyResourceManager {
	return rc.rom
}
func (rc *runtimeContext) SecretManager() secret_manager.SecretManager {
	return rc.sm
}
func (rc *runtimeContext) Extensions() context.Context {
	return rc.ext
}
