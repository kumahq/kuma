package runtime

import (
	"context"
	"sync"

	"github.com/kumahq/kuma/pkg/api-server/authn"
	api_server "github.com/kumahq/kuma/pkg/api-server/customization"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core/ca"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	managers_dataplane "github.com/kumahq/kuma/pkg/core/managers/apis/dataplane"
	managers_mesh "github.com/kumahq/kuma/pkg/core/managers/apis/mesh"
	resources_access "github.com/kumahq/kuma/pkg/core/resources/access"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/dns/resolver"
	dp_server "github.com/kumahq/kuma/pkg/dp-server/server"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/envoy/admin/access"
	"github.com/kumahq/kuma/pkg/events"
	kds_context "github.com/kumahq/kuma/pkg/kds/context"
	"github.com/kumahq/kuma/pkg/metrics"
	tokens_access "github.com/kumahq/kuma/pkg/tokens/builtin/access"
	zone_access "github.com/kumahq/kuma/pkg/tokens/builtin/zone/access"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"
	"github.com/kumahq/kuma/pkg/xds/secrets"
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
	DataSourceLoader() datasource.Loader
	ResourceManager() core_manager.ResourceManager
	ResourceStore() core_store.ResourceStore
	ReadOnlyResourceManager() core_manager.ReadOnlyResourceManager
	SecretStore() store.SecretStore
	ConfigStore() core_store.ResourceStore
	CaManagers() ca.Managers
	Extensions() context.Context
	DNSResolver() resolver.DNSResolver
	ConfigManager() config_manager.ConfigManager
	LeaderInfo() component.LeaderInfo
	LookupIP() lookup.LookupIPFunc
	EnvoyAdminClient() admin.EnvoyAdminClient
	Metrics() metrics.Metrics
	EventReaderFactory() events.ListenerFactory
	APIInstaller() api_server.APIInstaller
	XDSHooks() *xds_hooks.Hooks
	CAProvider() secrets.CaProvider
	DpServer() *dp_server.DpServer
	KDSContext() *kds_context.Context
	APIServerAuthenticator() authn.Authenticator
	ResourceValidators() ResourceValidators
	Access() Access
	// AppContext returns a context.Context which tracks the lifetime of the apps, it gets cancelled when the app is starting to shutdown.
	AppContext() context.Context
}

type Access struct {
	ResourceAccess       resources_access.ResourceAccess
	DataplaneTokenAccess tokens_access.DataplaneTokenAccess
	ZoneTokenAccess      zone_access.ZoneTokenAccess
	ConfigDumpAccess     access.ConfigDumpAccess
}

type ResourceValidators struct {
	Dataplane managers_dataplane.Validator
	Mesh      managers_mesh.MeshValidator
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
	dsl      datasource.Loader
	ext      context.Context
	dns      resolver.DNSResolver
	configm  config_manager.ConfigManager
	leadInfo component.LeaderInfo
	lif      lookup.LookupIPFunc
	eac      admin.EnvoyAdminClient
	metrics  metrics.Metrics
	erf      events.ListenerFactory
	apim     api_server.APIInstaller
	xdsh     *xds_hooks.Hooks
	cap      secrets.CaProvider
	dps      *dp_server.DpServer
	kdsctx   *kds_context.Context
	rv       ResourceValidators
	au       authn.Authenticator
	acc      Access
	appCtx   context.Context
}

func (rc *runtimeContext) Metrics() metrics.Metrics {
	return rc.metrics
}

func (rc *runtimeContext) EventReaderFactory() events.ListenerFactory {
	return rc.erf
}

func (rc *runtimeContext) CaManagers() ca.Managers {
	return rc.cam
}

func (rc *runtimeContext) Config() kuma_cp.Config {
	return rc.cfg
}

func (rc *runtimeContext) DataSourceLoader() datasource.Loader {
	return rc.dsl
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

func (rc *runtimeContext) DNSResolver() resolver.DNSResolver {
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

func (rc *runtimeContext) EnvoyAdminClient() admin.EnvoyAdminClient {
	return rc.eac
}

func (rc *runtimeContext) APIInstaller() api_server.APIInstaller {
	return rc.apim
}
func (rc *runtimeContext) DpServer() *dp_server.DpServer {
	return rc.dps
}

func (rc *runtimeContext) CAProvider() secrets.CaProvider {
	return rc.cap
}

func (rc *runtimeContext) XDSHooks() *xds_hooks.Hooks {
	return rc.xdsh
}

func (rc *runtimeContext) KDSContext() *kds_context.Context {
	return rc.kdsctx
}

func (rc *runtimeContext) ResourceValidators() ResourceValidators {
	return rc.rv
}

func (rc *runtimeContext) APIServerAuthenticator() authn.Authenticator {
	return rc.au
}

func (rc *runtimeContext) Access() Access {
	return rc.acc
}

func (rc *runtimeContext) AppContext() context.Context {
	return rc.appCtx
}
