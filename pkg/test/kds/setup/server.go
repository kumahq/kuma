package setup

import (
	"sync"
	"time"

	"github.com/Kong/kuma/pkg/kds/reconcile"

	. "github.com/onsi/gomega"

	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	kds_config "github.com/Kong/kuma/pkg/config/kds"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	"github.com/Kong/kuma/pkg/kds"
	kds_server "github.com/Kong/kuma/pkg/kds/server"
	test_grpc "github.com/Kong/kuma/pkg/test/grpc"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

type testRuntimeContext struct {
	runtime.Runtime
	rom        manager.ReadOnlyResourceManager
	cfg        kuma_cp.Config
	components []component.Component
}

func (t *testRuntimeContext) Config() kuma_cp.Config {
	return t.cfg
}

func (t *testRuntimeContext) ReadOnlyResourceManager() manager.ReadOnlyResourceManager {
	return t.rom
}

func (t *testRuntimeContext) Add(c ...component.Component) error {
	t.components = append(t.components, c...)
	return nil
}

func StartServer(store store.ResourceStore, wg *sync.WaitGroup) *test_grpc.MockServerStream {
	createServer := func(rt runtime.Runtime) kds_server.Server {
		log := core.Log
		hasher, cache := kds_server.NewXdsContext(log)
		generator := kds_server.NewSnapshotGenerator(rt, kds.SupportedTypes, reconcile.Any)
		versioner := kds_server.NewVersioner()
		reconciler := kds_server.NewReconciler(hasher, cache, generator, versioner)
		syncTracker := kds_server.NewSyncTracker(core.Log, reconciler, rt.Config().KDSServer.RefreshInterval)
		callbacks := util_xds.CallbacksChain{
			util_xds.LoggingCallbacks{Log: log},
			syncTracker,
		}
		return kds_server.NewServer(cache, callbacks, log)
	}

	srv := createServer(&testRuntimeContext{
		rom: manager.NewResourceManager(store),
		cfg: kuma_cp.Config{
			KDSServer: &kds_config.KumaDiscoveryServerConfig{
				RefreshInterval: 100 * time.Millisecond,
			},
		},
	})

	stream := test_grpc.MakeMockStream()
	go func() {
		err := srv.StreamKumaResources(stream)
		Expect(err).ToNot(HaveOccurred())
		wg.Done()
	}()
	return stream
}
