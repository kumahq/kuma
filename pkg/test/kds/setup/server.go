package setup

import (
	"sync"
	"time"

	. "github.com/onsi/gomega"

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
	test_grpc "github.com/kumahq/kuma/pkg/test/grpc"
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

func StartServer(store store.ResourceStore, wg *sync.WaitGroup, clusterID string, providedTypes []model.ResourceType, providedFilter reconcile.ResourceFilter, providedMapper reconcile.ResourceMapper) *test_grpc.MockServerStream {
	metrics, err := core_metrics.NewMetrics("Global")
	Expect(err).ToNot(HaveOccurred())
	rt := &testRuntimeContext{
		rom:     manager.NewResourceManager(store),
		cfg:     kuma_cp.Config{},
		metrics: metrics,
	}
	srv, err := kds_server.New(core.Log.WithName("kds").WithName(clusterID), rt, providedTypes, clusterID, 100*time.Millisecond, providedFilter, providedMapper, false)
	Expect(err).ToNot(HaveOccurred())
	stream := test_grpc.MakeMockStream()
	go func() {
		defer wg.Done()
		err := srv.StreamKumaResources(stream)
		Expect(err).ToNot(HaveOccurred())
	}()
	return stream
}
