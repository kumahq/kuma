package k8s

import (
	"context"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
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

type converterKey struct{}

type ConverterPredicate = func(core_model.Resource) bool

type Converter interface {
	ToKubernetesObject(core_model.Resource) (k8s_model.KubernetesObject, error)
	ToKubernetesList(core_model.ResourceList) (k8s_model.KubernetesList, error)
	ToCoreResource(obj k8s_model.KubernetesObject, out core_model.Resource) error
	ToCoreList(obj k8s_model.KubernetesList, out core_model.ResourceList, predicate ConverterPredicate) error
}

func NewResourceConverterContext(ctx context.Context, converter Converter) context.Context {
	return context.WithValue(ctx, converterKey{}, converter)
}

func FromResourceConverterContext(ctx context.Context) (converter Converter, ok bool) {
	converter, ok = ctx.Value(converterKey{}).(Converter)
	return
}
