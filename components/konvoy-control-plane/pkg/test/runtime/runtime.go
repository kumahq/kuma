package runtime

import (
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"

	konvoy_cp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-cp"
	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	bootstrap_universal "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/bootstrap/universal"
	resources_memory "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
)

var _ core_runtime.RuntimeInfo = TestRuntimeInfo{}

type TestRuntimeInfo struct {
	InstanceId string
}

func (i TestRuntimeInfo) GetInstanceId() string {
	return i.InstanceId
}

func BuilderFor(cfg konvoy_cp.Config) *core_runtime.Builder {
	return core_runtime.BuilderFor(cfg).
		WithComponentManager(bootstrap_universal.NewComponentManager()).
		WithResourceStore(resources_memory.NewStore()).
		WithXdsContext(core_xds.NewXdsContext())
}
