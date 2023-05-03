package k8s

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	util_k8s "github.com/kumahq/kuma/pkg/util/k8s"
)

type ResourceMapper interface {
	Map(resource model.Resource, namespace string) (k8s_model.KubernetesObject, error)
}
type ResourceMapperFunc func(resource model.Resource, namespace string) (k8s_model.KubernetesObject, error)

func (f ResourceMapperFunc) Map(resource model.Resource, namespace string) (k8s_model.KubernetesObject, error) {
	return f(resource, namespace)
}

func NewMapper(systemNamespace string, storeType store.StoreType, kubeFactory KubeFactory) ResourceMapper {
	return ResourceMapperFunc(func(resource model.Resource, namespace string) (k8s_model.KubernetesObject, error) {
		if storeType == store.KubernetesStore {
			// If we're on k8s we should just return the k8s object as is.
			res, err := (&SimpleConverter{KubeFactory: kubeFactory}).ToKubernetesObject(resource)
			if err != nil {
				return nil, err
			}
			if namespace != "" {
				res.SetNamespace(namespace)
			}
			return res, err
		}
		// otherwise we create a k8s resource with the universal object we have.
		rs, err := kubeFactory.NewObject(resource)
		if err != nil {
			return nil, err
		}
		if rs.Scope() == k8s_model.ScopeNamespace {
			name, ns, err := util_k8s.CoreNameToK8sName(resource.GetMeta().GetName())
			if err != nil {
				// if the original resource doesn't look like a kubernetes name ("name"."namespace")`, just use the default namespace.
				// this is in the case where someone calls this on a universal cluster. Exporting is a use-case for this.
				ns = systemNamespace
				name = resource.GetMeta().GetName()
			}
			if namespace != "" { // If the user is forcing the namespace accept it.
				ns = namespace
			}
			rs.SetName(name)
			rs.SetNamespace(ns)
		} else {
			rs.SetName(resource.GetMeta().GetName())
		}
		rs.SetMesh(resource.GetMeta().GetMesh())
		rs.SetCreationTimestamp(v1.NewTime(resource.GetMeta().GetCreationTime()))
		rs.SetSpec(resource.GetSpec())
		return rs, nil
	})
}
