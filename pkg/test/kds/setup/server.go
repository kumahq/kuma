package setup

import (
	"time"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/kds/reconcile"
	kds_server "github.com/kumahq/kuma/pkg/kds/server"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

type testRuntimeContext struct {
	runtime.Runtime
	rom        manager.ReadOnlyResourceManager
	cfg        kuma_cp.Config
	components []component.Component
	metrics    core_metrics.Metrics
}

func (t *testRuntimeContext) Config() kuma_cp.Config {
	return t.cfg
}

func (t *testRuntimeContext) ReadOnlyResourceManager() manager.ReadOnlyResourceManager {
	return t.rom
}

func (t *testRuntimeContext) Metrics() core_metrics.Metrics {
	return t.metrics
}

func (t *testRuntimeContext) Add(c ...component.Component) error {
	t.components = append(t.components, c...)
	return nil
}

func StartServer(store store.ResourceStore, clusterID string, providedTypes []model.ResourceType, providedFilter reconcile.ResourceFilter, providedMapper reconcile.ResourceMapper) (kds_server.Server, error) {
	metrics, err := core_metrics.NewMetrics("Global")
	if err != nil {
		return nil, err
	}
	rt := &testRuntimeContext{
		rom:     manager.NewResourceManager(store),
		cfg:     kuma_cp.Config{},
		metrics: metrics,
	}
	return kds_server.New(core.Log.WithName("kds").WithName(clusterID), rt, providedTypes, clusterID, 100*time.Millisecond, providedFilter, providedMapper, false, 1*time.Second)
}
