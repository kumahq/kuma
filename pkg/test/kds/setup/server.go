package setup

import (
	"sync"
	"time"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/kds/reconcile"
	kds_server "github.com/kumahq/kuma/pkg/kds/server"

	"github.com/kumahq/kuma/pkg/core/resources/model"

	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	kds_config "github.com/kumahq/kuma/pkg/config/kds"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	test_grpc "github.com/kumahq/kuma/pkg/test/grpc"
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

func StartServer(store store.ResourceStore, wg *sync.WaitGroup, clusterID string, providedTypes []model.ResourceType, providedFilter reconcile.ResourceFilter) *test_grpc.MockServerStream {
	rt := &testRuntimeContext{
		rom: manager.NewResourceManager(store),
		cfg: kuma_cp.Config{
			KDS: &kds_config.KdsConfig{
				Server: &kds_config.KdsServerConfig{
					RefreshInterval: 100 * time.Millisecond,
				},
			},
		},
	}
	srv, err := kds_server.New(core.Log, rt, providedTypes, clusterID, providedFilter)
	Expect(err).ToNot(HaveOccurred())
	stream := test_grpc.MakeMockStream()
	go func() {
		err := srv.StreamKumaResources(stream)
		Expect(err).ToNot(HaveOccurred())
		wg.Done()
	}()
	return stream
}
