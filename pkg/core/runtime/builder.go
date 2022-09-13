package runtime

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/api-server/authn"
	api_server "github.com/kumahq/kuma/pkg/api-server/customization"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/secrets/store"
	dp_server "github.com/kumahq/kuma/pkg/dp-server/server"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/events"
	kds_context "github.com/kumahq/kuma/pkg/kds/context"
	"github.com/kumahq/kuma/pkg/metrics"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"
	"github.com/kumahq/kuma/pkg/xds/secrets"
)

func NewRuntime(appCtx context.Context, cfg kuma_cp.Config) (Runtime, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, errors.Wrap(err, "could not get hostname")
	}
	suffix := core.NewUUID()[0:4]
	return &runtime{
		RuntimeContext: &runtimeContext{
			appCtx: appCtx,
			cfg:    cfg,
			ext:    context.Background(),
			cam:    core_ca.Managers{},
		},
		RuntimeInfo: &runtimeInfo{
			instanceId: fmt.Sprintf("%s-%s", hostname, suffix),
			startTime:  time.Now(),
		},
		Manager: nil,
	}, nil
}

// ValidateRuntime check that the runtime object has everything set as needed, this also blocks users from applying further options
func ValidateRuntime(r Runtime) error {
	if r.ResourceManager() == nil {
		return errors.Errorf("ComponentManager has not been configured")
	}
	if r.ResourceStore() == nil {
		return errors.Errorf("ResourceStore has not been configured")
	}
	if r.ResourceManager() == nil {
		return errors.Errorf("ResourceManager has not been configured")
	}
	if r.ReadOnlyResourceManager() == nil {
		return errors.Errorf("ReadOnlyResourceManager has not been configured")
	}
	if r.DataSourceLoader() == nil {
		return errors.Errorf("DataSourceLoader has not been configured")
	}
	if r.Extensions() == nil {
		return errors.Errorf("Extensions have been misconfigured")
	}
	if r.LeaderInfo() == nil {
		return errors.Errorf("LeaderInfo has not been configured")
	}
	if r.LookupIP() == nil {
		return errors.Errorf("LookupIP func has not been configured")
	}
	if r.EnvoyAdminClient() == nil {
		return errors.Errorf("EnvoyAdminClient has not been configured")
	}
	if r.Metrics() == nil {
		return errors.Errorf("Metrics has not been configured")
	}
	if r.EventReaderFactory() == nil {
		return errors.Errorf("EventReaderFactory has not been configured")
	}
	if r.APIInstaller() == nil {
		return errors.Errorf("APIInstaller has not been configured")
	}
	if r.CAProvider() == nil {
		return errors.Errorf("CAProvider has not been configured")
	}
	if r.DpServer() == nil {
		return errors.Errorf("DpServer has not been configured")
	}
	if r.KDSContext() == nil {
		return errors.Errorf("KDSContext has not been configured")
	}
	if r.ResourceValidators() == (ResourceValidators{}) {
		return errors.Errorf("ResourceValidators have not been configured")
	}
	if r.APIServerAuthenticator() == nil {
		return errors.Errorf("API Server Authenticator has not been configured")
	}
	if r.Access() == (Access{}) {
		return errors.Errorf("Access has not been configured")
	}
	r.(*runtime).validated = true
	return nil
}

func ApplyOpts(r Runtime, fns ...RuntimeOpt) error {
	for _, fn := range fns {
		if err := fn(r); err != nil {
			return err
		}
	}
	return nil
}

type RuntimeOpt func(r Runtime) error

func safeRuntimeOpt(fn func(r *runtimeContext)) RuntimeOpt {
	return func(r Runtime) error {
		if r.(*runtime).validated {
			return errors.New("Can't apply option as the object was already validated")
		}
		fn(r.(*runtime).RuntimeContext.(*runtimeContext))
		return nil
	}
}

func WithComponentManager(cm component.Manager) RuntimeOpt {
	return func(r Runtime) error {
		r.(*runtime).Manager = cm
		return nil
	}
}

func WithResourceStore(rs core_store.ResourceStore) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.rs = rs
	})
}

func WithSecretStore(ss store.SecretStore) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.ss = ss
	})
}

func WithConfigStore(cs core_store.ResourceStore) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.cs = cs
	})
}

func WithResourceManager(rm core_manager.CustomizableResourceManager) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.rm = rm
	})
}

func WithCustomizeResourceManager(fn func(manager core_manager.CustomizableResourceManager) error) RuntimeOpt {
	return func(r Runtime) error {
		if r.(*runtime).validated {
			return errors.New("Can't apply option as the object was already validated")
		}
		crm, ok := r.(*runtime).RuntimeContext.(*runtimeContext).rm.(core_manager.CustomizableResourceManager)
		if !ok {
			return errors.New("Can't customize resource manager that's not a CustomizableResourceManager")
		}
		return fn(crm)
	}
}

func WithReadOnlyResourceManager(rom core_manager.ReadOnlyResourceManager) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.rom = rom
	})
}

func WithCaManager(name string, cam core_ca.Manager) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.cam[name] = cam
	})
}

func WithDataSourceLoader(loader datasource.Loader) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.dsl = loader
	})
}

func WithExtensions(ext context.Context) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.ext = ext
	})
}

func WithConfigManager(configm config_manager.ConfigManager) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.configm = configm
	})
}

func WithLeaderInfo(leadInfo component.LeaderInfo) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.leadInfo = leadInfo
	})
}

func WithLookupIP(lif lookup.LookupIPFunc) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.lif = lif
	})
}

func WithEnvoyAdminClient(eac admin.EnvoyAdminClient) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.eac = eac
	})
}

func WithMetrics(metrics metrics.Metrics) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.metrics = metrics
	})
}

func WithEventReaderFactory(erf events.ListenerFactory) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.erf = erf
	})
}

func WithAPIManager(apim api_server.APIManager) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.apim = apim
	})
}

func WithAPIEndpoint(ws *restful.WebService) RuntimeOpt {
	return func(r Runtime) error {
		apim := r.(*runtime).RuntimeContext.(*runtimeContext).apim
		if apim == nil {
			return errors.New("No apimanager set")
		}
		apim.Add(ws)
		return nil
	}
}

func WithXDSResourceSetHook(xdsh xds_hooks.ResourceSetHook) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.xdsh = append(r.xdsh, xdsh)
	})
}

func WithCAProvider(cap secrets.CaProvider) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.cap = cap
	})
}

func WithDpServer(dps *dp_server.DpServer) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.dps = dps
	})
}

func WithResourceValidators(rv ResourceValidators) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.rv = rv
	})
}

func WithKDSContext(kdsctx *kds_context.Context) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.kdsctx = kdsctx
	})
}

func WithAPIServerAuthenticator(au authn.Authenticator) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.au = au
	})
}

func WithAccess(acc Access) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.acc = acc
	})
}

func WithExtraReportsFn(fn ExtraReportsFn) RuntimeOpt {
	return safeRuntimeOpt(func(r *runtimeContext) {
		r.extraReportsFn = fn
	})
}
