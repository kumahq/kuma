package k8s

import (
	"context"
	"time"

	"github.com/pkg/errors"

	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	config_store "github.com/Kong/kuma/pkg/core/config/store"
	config_model "github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	common_k8s "github.com/Kong/kuma/pkg/plugins/common/k8s"
)

var _ config_store.ConfigStore = &KubernetesStore{}

const (
	configMapKey = "config"
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
	opts := core_store.NewCreateOptions(fs...)
	cm := &kube_core.ConfigMap{
		TypeMeta: kube_meta.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: kube_meta.ObjectMeta{
			Name:      opts.Name,
			Namespace: s.namespace,
		},
		Immutable: nil,
		Data: map[string]string{
			configMapKey: r.Spec.Config,
		},
	}
	if err := s.client.Create(context.Background(), cm); err != nil {
		return err
	}
	r.SetMeta(&KubernetesMetaAdapter{cm.ObjectMeta})
	return nil
}
func (s *KubernetesStore) Update(ctx context.Context, r *config_model.ConfigResource, fs ...core_store.UpdateOptionsFunc) error {
	cm := &kube_core.ConfigMap{
		TypeMeta: kube_meta.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: kube_meta.ObjectMeta{
			Name:      r.GetMeta().GetName(),
			Namespace: s.namespace,
		},
		Immutable: nil,
		Data: map[string]string{
			configMapKey: r.Spec.Config,
		},
	}
	if err := s.client.Update(context.Background(), cm); err != nil {
		return err
	}
	r.SetMeta(&KubernetesMetaAdapter{cm.ObjectMeta})
	return nil
}
func (s *KubernetesStore) Delete(ctx context.Context, r *config_model.ConfigResource, fs ...core_store.DeleteOptionsFunc) error {
	opts := core_store.NewDeleteOptions(fs...)
	cm := &kube_core.ConfigMap{
		TypeMeta: kube_meta.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: kube_meta.ObjectMeta{
			Name:      opts.Name,
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
	r.SetMeta(&KubernetesMetaAdapter{cm.ObjectMeta})
	return nil
}
func (s *KubernetesStore) List(ctx context.Context, rs *config_model.ConfigResourceList, fs ...core_store.ListOptionsFunc) error {
	cmlist := &kube_core.ConfigMapList{}

	fields := kube_client.MatchingFields{
		"type": "ConfigMap",
	}
	if err := s.client.List(ctx, cmlist, kube_client.InNamespace(s.namespace), fields); err != nil {
		return errors.Wrap(err, "failed to list k8s internal config")
	}
	for _, cm := range cmlist.Items {
		rs.Items = append(rs.Items, &config_model.ConfigResource{
			Spec: system_proto.Config{
				Config: cm.Data[configMapKey],
			},
			Meta: &KubernetesMetaAdapter{cm.ObjectMeta},
		})
	}
	return nil
}

var _ core_model.ResourceMeta = &KubernetesMetaAdapter{}

type KubernetesMetaAdapter struct {
	kube_meta.ObjectMeta
}

func (m *KubernetesMetaAdapter) GetNameExtensions() core_model.ResourceNameExtensions {
	return common_k8s.ResourceNameExtensions(m.ObjectMeta.Namespace, m.ObjectMeta.Name)
}

func (m *KubernetesMetaAdapter) GetVersion() string {
	return m.ObjectMeta.GetResourceVersion()
}

func (m *KubernetesMetaAdapter) GetMesh() string {
	return ""
}

func (m *KubernetesMetaAdapter) GetCreationTime() time.Time {
	return m.GetObjectMeta().GetCreationTimestamp().Time
}

func (m *KubernetesMetaAdapter) GetModificationTime() time.Time {
	return m.GetObjectMeta().GetCreationTimestamp().Time
}
