package runtime

import (
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
)

var _ core_runtime.RuntimeInfo = TestRuntimeInfo{}

type TestRuntimeInfo struct {
	InstanceId string
}

func (i TestRuntimeInfo) GetInstanceId() string {
	return i.InstanceId
}
