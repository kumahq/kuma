package server

import (
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/kds/generator"
	mads_reconcile "github.com/Kong/kuma/pkg/mads/reconcile"
)

func NewSnapshotGenerator(rt core_runtime.Runtime) mads_reconcile.SnapshotGenerator {
	return generator.NewSnapshotGenerator(rt.ReadOnlyResourceManager())
}
