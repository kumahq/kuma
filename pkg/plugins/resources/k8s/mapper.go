package k8s

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
)

type ResourceMapperFunc func(resource model.Resource, namespace string) (k8s_model.KubernetesObject, error)

// NewKubernetesMapper creates a ResourceMapper that returns the k8s object as is. This is meant to be used when the underlying store is kubernetes
func NewKubernetesMapper(kubeFactory KubeFactory) ResourceMapperFunc {
	return func(resource model.Resource, namespace string) (k8s_model.KubernetesObject, error) {
		res, err := (&SimpleConverter{KubeFactory: kubeFactory}).ToKubernetesObject(resource)
		if err != nil {
			return nil, err
		}
		if namespace != "" {
			res.SetNamespace(namespace)
		}
		return res, err
	}
}

// NewInferenceMapper creates a ResourceMapper that infers a k8s resource from the core_model. Extract namespace from the name if necessary.
// This mostly useful when the underlying store is not kubernetes but you want to show what a kubernetes version of the policy would be like (in global for example).
func NewInferenceMapper(systemNamespace string, kubeFactory KubeFactory) ResourceMapperFunc {
	return func(resource model.Resource, namespace string) (k8s_model.KubernetesObject, error) {
		rs, err := kubeFactory.NewObject(resource)
		if err != nil {
			return nil, err
		}
		if rs.Scope() == k8s_model.ScopeNamespace {
			if namespace != "" { // If the user is forcing the namespace accept it.
				rs.SetNamespace(namespace)
			} else {
				rs.SetNamespace(systemNamespace)
			}
		}
		rs.SetName(resource.GetMeta().GetName())
		rs.SetMesh(resource.GetMeta().GetMesh())
		rs.SetCreationTimestamp(v1.NewTime(resource.GetMeta().GetCreationTime()))
		rs.SetSpec(resource.GetSpec())
		if resource.Descriptor().HasStatus {
			if err := rs.SetStatus(resource.GetStatus()); err != nil {
				return nil, err
			}
		}
		return rs, nil
	}
}
