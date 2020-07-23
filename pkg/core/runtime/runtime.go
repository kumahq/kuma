package runtime

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/secrets/store"

	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"

	"github.com/kumahq/kuma/pkg/zones/poller"

	"github.com/kumahq/kuma/pkg/dns"

	"github.com/kumahq/kuma/pkg/core/ca"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
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
	ResourceStore() core_store.ResourceStore
	ReadOnlyResourceManager() core_manager.ReadOnlyResourceManager
	SecretStore() store.SecretStore
	CaManagers() ca.Managers
	Extensions() context.Context
	DNSResolver() dns.DNSResolver
	Zones() poller.ZoneStatusPoller
	ConfigManager() config_manager.ConfigManager
	LeaderInfo() component.LeaderInfo
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
	cfg      kuma_cp.Config
	rm       core_manager.ResourceManager
	rs       core_store.ResourceStore
	ss       store.SecretStore
	rom      core_manager.ReadOnlyResourceManager
	cam      ca.Managers
	xds      core_xds.XdsContext
	ext      context.Context
	dns      dns.DNSResolver
	zones    poller.ZoneStatusPoller
	configm  config_manager.ConfigManager
	leadInfo component.LeaderInfo
}

func (rc *runtimeContext) CaManagers() ca.Managers {
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
func (rc *runtimeContext) ResourceStore() core_store.ResourceStore {
	return rc.rs
}
func (rc *runtimeContext) SecretStore() store.SecretStore {
	return rc.ss
}
func (rc *runtimeContext) ReadOnlyResourceManager() core_manager.ReadOnlyResourceManager {
	return rc.rom
}
func (rc *runtimeContext) Extensions() context.Context {
	return rc.ext
}

func (rc *runtimeContext) DNSResolver() dns.DNSResolver {
	return rc.dns
}

func (rc *runtimeContext) Zones() poller.ZoneStatusPoller {
	return rc.zones
}

func (rc *runtimeContext) ConfigManager() config_manager.ConfigManager {
	return rc.configm
}

func (rc *runtimeContext) LeaderInfo() component.LeaderInfo {
	return rc.leadInfo
}
