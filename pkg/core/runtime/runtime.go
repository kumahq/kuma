package runtime

import (
	"context"
	"sync"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/metrics"

	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"

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
	SetClusterId(clusterId string)
	GetClusterId() string
}

type RuntimeContext interface {
	Config() kuma_cp.Config
	XDS() core_xds.XdsContext
	ResourceManager() core_manager.ResourceManager
	ResourceStore() core_store.ResourceStore
	ReadOnlyResourceManager() core_manager.ReadOnlyResourceManager
	SecretStore() store.SecretStore
	ConfigStore() core_store.ResourceStore
	CaManagers() ca.Managers
	Extensions() context.Context
	DNSResolver() dns.DNSResolver
	ConfigManager() config_manager.ConfigManager
	LeaderInfo() component.LeaderInfo
	LookupIP() lookup.LookupIPFunc
	Metrics() metrics.Metrics
	EventReaderFactory() events.ReaderFactory
}

var _ Runtime = &runtime{}

type runtime struct {
	RuntimeInfo
	RuntimeContext
	component.Manager
}

var _ RuntimeInfo = &runtimeInfo{}

type runtimeInfo struct {
	mtx sync.RWMutex

	instanceId string
	clusterId  string
}

func (i *runtimeInfo) GetInstanceId() string {
	return i.instanceId
}

func (i *runtimeInfo) SetClusterId(clusterId string) {
	i.mtx.Lock()
	defer i.mtx.Unlock()
	i.clusterId = clusterId
}

func (i *runtimeInfo) GetClusterId() string {
	i.mtx.RLock()
	defer i.mtx.RUnlock()
	return i.clusterId
}

var _ RuntimeContext = &runtimeContext{}

type runtimeContext struct {
	cfg      kuma_cp.Config
	rm       core_manager.ResourceManager
	rs       core_store.ResourceStore
	ss       store.SecretStore
	cs       core_store.ResourceStore
	rom      core_manager.ReadOnlyResourceManager
	cam      ca.Managers
	xds      core_xds.XdsContext
	ext      context.Context
	dns      dns.DNSResolver
	configm  config_manager.ConfigManager
	leadInfo component.LeaderInfo
	lif      lookup.LookupIPFunc
	metrics  metrics.Metrics
	erf      events.ReaderFactory
}

func (rc *runtimeContext) Metrics() metrics.Metrics {
	return rc.metrics
}

func (rc *runtimeContext) EventReaderFactory() events.ReaderFactory {
	return rc.erf
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

func (rc *runtimeContext) ConfigStore() core_store.ResourceStore {
	return rc.cs
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

func (rc *runtimeContext) ConfigManager() config_manager.ConfigManager {
	return rc.configm
}

func (rc *runtimeContext) LeaderInfo() component.LeaderInfo {
	return rc.leadInfo
}

func (rc *runtimeContext) LookupIP() lookup.LookupIPFunc {
	return rc.lif
}
