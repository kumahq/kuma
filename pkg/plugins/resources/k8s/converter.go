package k8s

import (
	"fmt"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	k8s_extensions "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ k8s_extensions.Converter = &SimpleConverter{}

type SimpleConverter struct {
	KubeFactory KubeFactory
}

func (c *SimpleConverter) ToKubernetesObject(r core_model.Resource) (k8s_model.KubernetesObject, error) {
	obj, err := c.KubeFactory.NewObject(r)
	if err != nil {
		return nil, err
	}
	spec, err := util_proto.ToMap(r.GetSpec())
	if err != nil {
		return nil, err
	}
	obj.SetSpec(spec)
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
	return util_proto.FromMap(obj.GetSpec(), out.GetSpec())
}

func (c *SimpleConverter) ToCoreList(in k8s_model.KubernetesList, out core_model.ResourceList, predicate k8s_extensions.ConverterPredicate) error {
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
