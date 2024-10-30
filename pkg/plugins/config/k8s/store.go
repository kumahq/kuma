package k8s

import (
	"context"
	"maps"
	"time"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	config_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	common_k8s "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
)

var _ core_store.ResourceStore = &KubernetesStore{}

const (
	configMapKey = "config"
)

type KubernetesStore struct {
	client kube_client.Client
	// Namespace to store ConfigMaps in, e.g. namespace where Control Plane is installed to
	namespace string
	converter common_k8s.Converter
	scheme    *kube_runtime.Scheme
}

func NewStore(client kube_client.Client, namespace string, scheme *kube_runtime.Scheme, converter common_k8s.Converter) (core_store.ResourceStore, error) {
	return &KubernetesStore{
		client:    client,
		namespace: namespace,
		converter: converter,
		scheme:    scheme,
	}, nil
}

func (s *KubernetesStore) Create(ctx context.Context, r core_model.Resource, fs ...core_store.CreateOptionsFunc) error {
	configRes, ok := r.(*config_model.ConfigResource)
	if !ok {
		return newInvalidTypeError()
	}
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
			configMapKey: configRes.Spec.Config,
		},
	}

	labels, annotations := k8s.SplitLabelsAndAnnotations(opts.Labels, cm.GetAnnotations())
	cm.GetObjectMeta().SetLabels(labels)
	cm.GetObjectMeta().SetAnnotations(annotations)

	if opts.Owner != nil {
		k8sOwner, err := s.converter.ToKubernetesObject(opts.Owner)
		if err != nil {
			return errors.Wrap(err, "failed to convert core model into k8s counterpart")
		}
		if err := controllerutil.SetOwnerReference(k8sOwner, cm, s.scheme); err != nil {
			return errors.Wrap(err, "failed to set owner reference for object")
		}
	}
	if err := s.client.Create(ctx, cm); err != nil {
		return err
	}
	r.SetMeta(&KubernetesMetaAdapter{cm.ObjectMeta})
	return nil
}

func (s *KubernetesStore) Update(ctx context.Context, r core_model.Resource, fs ...core_store.UpdateOptionsFunc) error {
	configRes, ok := r.(*config_model.ConfigResource)
	if !ok {
		return newInvalidTypeError()
	}
	opts := core_store.NewUpdateOptions(fs...)
	cm := &kube_core.ConfigMap{
		TypeMeta: kube_meta.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: r.GetMeta().(*KubernetesMetaAdapter).ObjectMeta,
		Immutable:  nil,
		Data: map[string]string{
			configMapKey: configRes.Spec.Config,
		},
	}

	updateLabels := cm.GetLabels()
	if opts.ModifyLabels {
		updateLabels = opts.Labels
	}

	labels, annotations := k8s.SplitLabelsAndAnnotations(updateLabels, cm.GetAnnotations())
	cm.GetObjectMeta().SetLabels(labels)
	cm.GetObjectMeta().SetAnnotations(annotations)

	if err := s.client.Update(ctx, cm); err != nil {
		if kube_apierrs.IsConflict(err) {
			return core_store.ErrorResourceConflict(r.Descriptor().Name, r.GetMeta().GetName(), r.GetMeta().GetMesh())
		}
		return errors.Wrap(err, "failed to update k8s resource")
	}
	r.SetMeta(&KubernetesMetaAdapter{cm.ObjectMeta})
	return nil
}

func (s *KubernetesStore) Delete(ctx context.Context, r core_model.Resource, fs ...core_store.DeleteOptionsFunc) error {
	configRes, ok := r.(*config_model.ConfigResource)
	if !ok {
		return newInvalidTypeError()
	}
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
			configMapKey: configRes.Spec.Config,
		},
	}
	return s.client.Delete(ctx, cm)
}

func (s *KubernetesStore) Get(ctx context.Context, r core_model.Resource, fs ...core_store.GetOptionsFunc) error {
	configRes, ok := r.(*config_model.ConfigResource)
	if !ok {
		return newInvalidTypeError()
	}
	opts := core_store.NewGetOptions(fs...)
	cm := &kube_core.ConfigMap{}
	if err := s.client.Get(ctx, kube_client.ObjectKey{Namespace: s.namespace, Name: opts.Name}, cm); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return core_store.ErrorResourceNotFound(r.Descriptor().Name, opts.Name, opts.Mesh)
		}
		return errors.Wrap(err, "failed to get k8s Config")
	}
	configRes.Spec.Config = cm.Data[configMapKey]
	r.SetMeta(&KubernetesMetaAdapter{cm.ObjectMeta})
	return nil
}

func (s *KubernetesStore) List(ctx context.Context, rs core_model.ResourceList, fs ...core_store.ListOptionsFunc) error {
	configRes, ok := rs.(*config_model.ConfigResourceList)
	if !ok {
		return newInvalidTypeError()
	}
	cmlist := &kube_core.ConfigMapList{}

	if err := s.client.List(ctx, cmlist, kube_client.InNamespace(s.namespace)); err != nil {
		return errors.Wrap(err, "failed to list k8s internal config")
	}
	for _, cm := range cmlist.Items {
		configRes.Items = append(configRes.Items, &config_model.ConfigResource{
			Spec: &system_proto.Config{
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

func (m *KubernetesMetaAdapter) GetLabels() map[string]string {
	labels := maps.Clone(m.GetObjectMeta().GetLabels())
	if labels == nil {
		labels = map[string]string{}
	}
	if displayName, ok := m.GetObjectMeta().GetAnnotations()[v1alpha1.DisplayName]; ok {
		labels[v1alpha1.DisplayName] = displayName
	} else {
		labels[v1alpha1.DisplayName] = m.GetObjectMeta().GetName()
	}
	return labels
}

func newInvalidTypeError() error {
	return errors.New("resource has a wrong type")
}
