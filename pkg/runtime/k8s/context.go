package k8s

import (
	"context"

	kube_ctrl "sigs.k8s.io/controller-runtime"
)

type managerKey struct{}

func NewManagerContext(ctx context.Context, manager kube_ctrl.Manager) context.Context {
	return context.WithValue(ctx, managerKey{}, manager)
}

func FromManagerContext(ctx context.Context) (manager kube_ctrl.Manager, ok bool) {
	manager, ok = ctx.Value(managerKey{}).(kube_ctrl.Manager)
	return
}
