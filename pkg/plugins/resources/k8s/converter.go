package k8s

import (
	"fmt"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

var _ k8s_common.Converter = &SimpleConverter{}

type SimpleConverter struct {
	KubeFactory KubeFactory
}

func NewSimpleConverter() k8s_common.Converter {
	return &SimpleConverter{
		KubeFactory: NewSimpleKubeFactory(),
	}
}

func NewSimpleKubeFactory() KubeFactory {
	return &SimpleKubeFactory{
		KubeTypes: registry.Global(),
	}
}

func (c *SimpleConverter) ToKubernetesObject(r core_model.Resource) (k8s_model.KubernetesObject, error) {
	obj, err := c.KubeFactory.NewObject(r)
	if err != nil {
		return nil, err
	}
	obj.SetSpec(r.GetSpec())
	if r.GetMeta() != nil {
		if adapter, ok := r.GetMeta().(*KubernetesMetaAdapter); ok {
			obj.SetMesh(adapter.Mesh)
			obj.SetObjectMeta(&adapter.ObjectMeta)
		} else {
			return nil, fmt.Errorf("meta has unexpected type: %#v", r.GetMeta())
		}
	}
	return obj, nil
}

func (c *SimpleConverter) ToKubernetesList(rl core_model.ResourceList) (k8s_model.KubernetesList, error) {
	return c.KubeFactory.NewList(rl)
}

func (c *SimpleConverter) ToCoreResource(obj k8s_model.KubernetesObject, out core_model.Resource) error {
	out.SetMeta(&KubernetesMetaAdapter{ObjectMeta: *obj.GetObjectMeta(), Mesh: obj.GetMesh()})
	err := out.SetSpec(obj.GetSpec())
	return err
}

func (c *SimpleConverter) ToCoreList(in k8s_model.KubernetesList, out core_model.ResourceList, predicate k8s_common.ConverterPredicate) error {
	for _, o := range in.GetItems() {
		r := out.NewItem()
		if err := c.ToCoreResource(o, r); err != nil {
			return err
		}
		if predicate(r) {
			_ = out.AddItem(r)
		}
	}
	return nil
}
