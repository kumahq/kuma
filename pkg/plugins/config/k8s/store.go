package k8s

import (
	"context"

	"github.com/pkg/errors"

	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	config_store "github.com/Kong/kuma/pkg/core/config/store"
	config_model "github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
)

var _ config_store.ConfigStore = &KubernetesStore{}

const (
	configMapName = "kuma-internal-config"
	configMapKey  = "config"
)

type KubernetesStore struct {
	client kube_client.Client
	// Namespace to store ConfigMaps in, e.g. namespace where Control Plane is installed to
	namespace string
}

func NewStore(client kube_client.Client, namespace string) (config_store.ConfigStore, error) {
	return &KubernetesStore{
		client:    client,
		namespace: namespace,
	}, nil
}

func (s *KubernetesStore) Create(ctx context.Context, r *config_model.ConfigResource, fs ...core_store.CreateOptionsFunc) error {
	cm := &kube_core.ConfigMap{
		TypeMeta: kube_meta.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: kube_meta.ObjectMeta{
			Name:      configMapName,
			Namespace: s.namespace,
		},
		Immutable: nil,
		Data: map[string]string{
			configMapKey: r.Spec.Config,
		},
	}
	return s.client.Create(context.Background(), cm)
}
func (s *KubernetesStore) Update(ctx context.Context, r *config_model.ConfigResource, fs ...core_store.UpdateOptionsFunc) error {
	cm := &kube_core.ConfigMap{
		TypeMeta: kube_meta.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: kube_meta.ObjectMeta{
			Name:      configMapName,
			Namespace: s.namespace,
		},
		Immutable: nil,
		Data: map[string]string{
			configMapKey: r.Spec.Config,
		},
	}
	return s.client.Update(context.Background(), cm)
}
func (s *KubernetesStore) Delete(ctx context.Context, r *config_model.ConfigResource, fs ...core_store.DeleteOptionsFunc) error {
	cm := &kube_core.ConfigMap{
		TypeMeta: kube_meta.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: kube_meta.ObjectMeta{
			Name:      configMapName,
			Namespace: s.namespace,
		},
		Immutable: nil,
		Data: map[string]string{
			configMapKey: r.Spec.Config,
		},
	}
	return s.client.Delete(context.Background(), cm)
}
func (s *KubernetesStore) Get(ctx context.Context, r *config_model.ConfigResource, fs ...core_store.GetOptionsFunc) error {
	opts := core_store.NewGetOptions(fs...)
	cm := &kube_core.ConfigMap{}
	if err := s.client.Get(ctx, kube_client.ObjectKey{Namespace: s.namespace, Name: opts.Name}, cm); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return core_store.ErrorResourceNotFound(r.GetType(), opts.Name, opts.Mesh)
		}
		return errors.Wrap(err, "failed to get k8s Config")
	}
	r.Spec.Config = cm.Data[configMapKey]
	return nil
}
func (s *KubernetesStore) List(ctx context.Context, rs *config_model.ConfigResourceList, fs ...core_store.ListOptionsFunc) error {
	cmlist := &kube_core.ConfigMapList{}

	fields := kube_client.MatchingFields{
		"type": "ConfigMap",
		"name": configMapName,
	}
	if err := s.client.List(ctx, cmlist, kube_client.InNamespace(s.namespace), fields); err != nil {
		return errors.Wrap(err, "failed to list k8s internal config")
	}
	for _, cm := range cmlist.Items {
		rs.Items = append(rs.Items, &config_model.ConfigResource{
			Spec: system_proto.Config{
				Config: cm.Data[configMapKey],
			},
		})
	}
	return nil
}
