package defaults

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/sethvargo/go-retry"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/core/user"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
)

var log = core.Log.WithName("defaults")

type EnsureDefaultFunc = func(ctx context.Context, resManager core_manager.ResourceManager, logger logr.Logger, cfg kuma_cp.Config) error

var EnsureDefaultFuncs = []EnsureDefaultFunc{
	EnsureEnvoyAdminCaExists,
	EnsureOnlyOneZoneExists,
	EnsureDefaultMeshExists,
	EnsureZoneTokenSigningKeyExists,
	EnsureHostnameGeneratorExists,
}

func Setup(runtime runtime.Runtime) error {
	if runtime.Config().Defaults.Mode == kuma_cp.ModeNone {
		log.V(1).Info("skipping default tenant resources because KUMA_DEFAULTS_MODE is set to none")
		return nil
	}

	if runtime.Config().Defaults.SkipTenantResources {
		log.V(1).Info("skipping default tenant resources because KUMA_DEFAULTS_SKIP_TENANT_RESOURCES is set to true")
		return nil
	}

	return runtime.Add(&DefaultComponent{
		Extensions: runtime.Extensions(),
		Funcs:      EnsureDefaultFuncs,
		ResManager: runtime.ResourceManager(),
		CpConfig:   runtime.Config(),
	})
}

type DefaultComponent struct {
	Extensions context.Context
	Funcs      []EnsureDefaultFunc
	ResManager core_manager.ResourceManager
	CpConfig   kuma_cp.Config
}

var _ component.Component = &DefaultComponent{}

func (e *DefaultComponent) Start(stop <-chan struct{}) error {
	ctx, cancelFn := context.WithCancel(user.Ctx(context.Background(), user.ControlPlane))
	defer cancelFn()
	logger := kuma_log.AddFieldsFromCtx(log, ctx, e.Extensions)
	errChan := make(chan error)
	go func() {
		errChan <- retry.Do(ctx, retry.WithMaxDuration(10*time.Minute, retry.NewConstant(5*time.Second)), func(ctx context.Context) error {
			for _, fn := range e.Funcs {
				if err := fn(ctx, e.ResManager, logger, e.CpConfig); err != nil {
					logger.V(1).Info("could not ensure default resources. Retrying.", "err", err)
					return retry.RetryableError(err)
				}
			}
			return nil
		})
	}()
	select {
	case <-stop:
		return nil
	case err := <-errChan:
		return err
	}
}

func (e DefaultComponent) NeedLeaderElection() bool {
	return true
}

func EnsureZoneTokenSigningKeyExists(ctx context.Context, resManager core_manager.ResourceManager, logger logr.Logger, cfg kuma_cp.Config) error {
	if cfg.IsFederatedZoneCP() {
		return nil
	}
	return tokens.EnsureDefaultSigningKeyExist(zone.SigningKeyPrefix, ctx, resManager, logger)
}
