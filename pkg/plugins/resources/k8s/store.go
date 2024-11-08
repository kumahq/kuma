package k8s

import (
	"context"
	"maps"
	"strings"
	"time"

	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	util_k8s "github.com/kumahq/kuma/pkg/util/k8s"
)

func typeIsUnregistered(err error) bool {
	var typeErr *k8s_registry.UnknownTypeError
	return errors.As(err, &typeErr)
}

var _ store.ResourceStore = &KubernetesStore{}

type KubernetesStore struct {
	Client    kube_client.Client
	Converter k8s_common.Converter
	Scheme    *kube_runtime.Scheme
}

func NewStore(client kube_client.Client, scheme *kube_runtime.Scheme, converter k8s_common.Converter) (store.ResourceStore, error) {
	return &KubernetesStore{
		Client:    client,
		Converter: converter,
		Scheme:    scheme,
	}, nil
}

func (s *KubernetesStore) Create(ctx context.Context, r core_model.Resource, fs ...store.CreateOptionsFunc) error {
	opts := store.NewCreateOptions(fs...)
	obj, err := s.Converter.ToKubernetesObject(r)
	if err != nil {
		if typeIsUnregistered(err) {
			return errors.Errorf("cannot create instance of unregistered type %q", r.Descriptor().Name)
		}
		return errors.Wrap(err, "failed to convert core model into k8s counterpart")
	}
	name, namespace, err := k8sNameNamespace(opts.Name, obj.Scope())
	if err != nil {
		return err
	}

	labels, annotations := SplitLabelsAndAnnotations(opts.Labels, obj.GetAnnotations())
	obj.GetObjectMeta().SetLabels(labels)
	obj.GetObjectMeta().SetAnnotations(annotations)
	obj.SetMesh(opts.Mesh)
	obj.GetObjectMeta().SetName(name)
	obj.GetObjectMeta().SetNamespace(namespace)

	if opts.Owner != nil {
		k8sOwner, err := s.Converter.ToKubernetesObject(opts.Owner)
		if err != nil {
			return errors.Wrap(err, "failed to convert core model into k8s counterpart")
		}
		if err := controllerutil.SetOwnerReference(k8sOwner, obj, s.Scheme); err != nil {
			return errors.Wrap(err, "failed to set owner reference for object")
		}
	}

	if err := s.Client.Create(ctx, obj); err != nil {
		if kube_apierrs.IsAlreadyExists(err) {
			return store.ErrorResourceAlreadyExists(r.Descriptor().Name, opts.Name, opts.Mesh)
		}
		return errors.Wrap(err, "failed to create k8s resource")
	}
	err = s.Converter.ToCoreResource(obj, r)
	if err != nil {
		return errors.Wrap(err, "failed to convert k8s model into core counterpart")
	}
	return nil
}

func (s *KubernetesStore) Update(ctx context.Context, r core_model.Resource, fs ...store.UpdateOptionsFunc) error {
	opts := store.NewUpdateOptions(fs...)

	obj, err := s.Converter.ToKubernetesObject(r)
	if err != nil {
		if typeIsUnregistered(err) {
			return errors.Errorf("cannot update instance of unregistered type %q", r.Descriptor().Name)
		}
		return errors.Wrapf(err, "failed to convert core model of type %s into k8s counterpart", r.Descriptor().Name)
	}

	updateLabels := r.GetMeta().GetLabels()
	if opts.ModifyLabels {
		updateLabels = opts.Labels
	}
	labels, annotations := SplitLabelsAndAnnotations(updateLabels, obj.GetAnnotations())

	obj.GetObjectMeta().SetLabels(labels)
	obj.GetObjectMeta().SetAnnotations(annotations)
	obj.SetMesh(r.GetMeta().GetMesh())

	if err := s.Client.Update(ctx, obj); err != nil {
		if kube_apierrs.IsConflict(err) {
			return store.ErrorResourceConflict(r.Descriptor().Name, r.GetMeta().GetName(), r.GetMeta().GetMesh())
		}
		return errors.Wrap(err, "failed to update k8s resource")
	}
	err = s.Converter.ToCoreResource(obj, r)
	if err != nil {
		return errors.Wrap(err, "failed to convert k8s model into core counterpart")
	}
	return nil
}

func (s *KubernetesStore) Delete(ctx context.Context, r core_model.Resource, fs ...store.DeleteOptionsFunc) error {
	opts := store.NewDeleteOptions(fs...)

	// get object and validate mesh
	if err := s.Get(ctx, r, store.GetByKey(opts.Name, opts.Mesh)); err != nil {
		return err
	}

	obj, err := s.Converter.ToKubernetesObject(r)
	if err != nil {
		// Unregistered types can't exist in the first place, so deletion would automatically succeed.
		if typeIsUnregistered(err) {
			return nil
		}
		return errors.Wrapf(err, "failed to convert core model of type %s into k8s counterpart", r.Descriptor().Name)
	}

	name, namespace, err := k8sNameNamespace(opts.Name, obj.Scope())
	if err != nil {
		return err
	}
	obj.GetObjectMeta().SetName(name)
	obj.GetObjectMeta().SetNamespace(namespace)
	if err := s.Client.Delete(ctx, obj); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return nil
		}
		return errors.Wrap(err, "failed to delete k8s resource")
	}
	return nil
}

func (s *KubernetesStore) Get(ctx context.Context, r core_model.Resource, fs ...store.GetOptionsFunc) error {
	opts := store.NewGetOptions(fs...)
	obj, err := s.Converter.ToKubernetesObject(r)
	if err != nil {
		if typeIsUnregistered(err) {
			return store.ErrorResourceNotFound(r.Descriptor().Name, opts.Name, opts.Mesh)
		}
		return errors.Wrapf(err, "failed to convert core model of type %s into k8s counterpart", r.Descriptor().Name)
	}
	name, namespace, err := k8sNameNamespace(opts.Name, obj.Scope())
	if err != nil {
		return err
	}
	if err := s.Client.Get(ctx, kube_client.ObjectKey{Namespace: namespace, Name: name}, obj); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return store.ErrorResourceNotFound(r.Descriptor().Name, opts.Name, opts.Mesh)
		}
		return errors.Wrap(err, "failed to get k8s resource")
	}
	if err := s.Converter.ToCoreResource(obj, r); err != nil {
		return errors.Wrap(err, "failed to convert k8s model into core counterpart")
	}
	if opts.Version != "" && r.GetMeta().GetVersion() != opts.Version {
		return store.ErrorResourceConflict(r.Descriptor().Name, opts.Name, opts.Mesh)
	}
	if r.GetMeta().GetMesh() != opts.Mesh {
		return store.ErrorResourceNotFound(r.Descriptor().Name, opts.Name, opts.Mesh)
	}
	return nil
}

func (s *KubernetesStore) List(ctx context.Context, rs core_model.ResourceList, fs ...store.ListOptionsFunc) error {
	opts := store.NewListOptions(fs...)
	obj, err := s.Converter.ToKubernetesList(rs)
	if err != nil {
		if typeIsUnregistered(err) {
			return nil
		}
		return errors.Wrapf(err, "failed to convert core list model of type %s into k8s counterpart", rs.GetItemType())
	}
	if err := s.Client.List(ctx, obj); err != nil {
		return errors.Wrap(err, "failed to list k8s resources")
	}
	predicate := func(r core_model.Resource) bool {
		if opts.Mesh != "" && r.GetMeta().GetMesh() != opts.Mesh {
			return false
		}
		if opts.NameContains != "" && !strings.Contains(r.GetMeta().GetName(), opts.NameContains) {
			return false
		}
		return true
	}
	fullList, err := registry.Global().NewList(rs.GetItemType())
	if err != nil {
		return err
	}
	if err := s.Converter.ToCoreList(obj, fullList, predicate); err != nil {
		return errors.Wrap(err, "failed to convert k8s model into core counterpart")
	}

	for _, item := range fullList.GetItems() {
		_ = rs.AddItem(item)
	}

	rs.GetPagination().SetTotal(uint32(len(fullList.GetItems())))
	return nil
}

func k8sNameNamespace(coreName string, scope k8s_model.Scope) (string, string, error) {
	if coreName == "" {
		return "", "", store.PreconditionFormatError("name can't be empty")
	}
	switch scope {
	case k8s_model.ScopeCluster:
		return coreName, "", nil
	case k8s_model.ScopeNamespace:
		name, ns, err := util_k8s.CoreNameToK8sName(coreName)
		if err != nil {
			return "", "", store.PreconditionFormatError(err.Error())
		}
		return name, ns, nil
	default:
		return "", "", errors.Errorf("unknown scope %s", scope)
	}
}

// Kuma resource labels are generally stored on Kubernetes as labels, except "kuma.io/display-name".
// We store it as an annotation because the resource name on k8s is limited by 253 and the label value is limited by 63.
func SplitLabelsAndAnnotations(coreLabels map[string]string, currentAnnotations map[string]string) (map[string]string, map[string]string) {
	labels := maps.Clone(coreLabels)
	annotations := maps.Clone(currentAnnotations)
	if annotations == nil {
		annotations = map[string]string{}
	}
	if v, ok := labels[v1alpha1.DisplayName]; ok {
		annotations[v1alpha1.DisplayName] = v
		delete(labels, v1alpha1.DisplayName)
	}
	return labels, annotations
}

var _ core_model.ResourceMeta = &KubernetesMetaAdapter{}

type KubernetesMetaAdapter struct {
	kube_meta.ObjectMeta
	Mesh string
}

func (m *KubernetesMetaAdapter) GetName() string {
	if m.Namespace == "" { // it's cluster scoped object
		return m.ObjectMeta.Name
	}
	return util_k8s.K8sNamespacedNameToCoreName(m.ObjectMeta.Name, m.ObjectMeta.Namespace)
}

func (m *KubernetesMetaAdapter) GetNameExtensions() core_model.ResourceNameExtensions {
	return k8s_common.ResourceNameExtensions(m.ObjectMeta.Namespace, m.ObjectMeta.Name)
}

func (m *KubernetesMetaAdapter) GetVersion() string {
	return m.ObjectMeta.GetResourceVersion()
}

func (m *KubernetesMetaAdapter) GetMesh() string {
	return m.Mesh
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
	if _, ok := labels[v1alpha1.KubeNamespaceTag]; !ok && m.Namespace != "" {
		labels[v1alpha1.KubeNamespaceTag] = m.Namespace
	}

	return labels
}

type KubeFactory interface {
	NewObject(r core_model.Resource) (k8s_model.KubernetesObject, error)
	NewList(rl core_model.ResourceList) (k8s_model.KubernetesList, error)
}

var _ KubeFactory = &SimpleKubeFactory{}

type SimpleKubeFactory struct {
	KubeTypes k8s_registry.TypeRegistry
}

func (f *SimpleKubeFactory) NewObject(r core_model.Resource) (k8s_model.KubernetesObject, error) {
	return f.KubeTypes.NewObject(r.GetSpec())
}

func (f *SimpleKubeFactory) NewList(rl core_model.ResourceList) (k8s_model.KubernetesList, error) {
	return f.KubeTypes.NewList(rl.NewItem().GetSpec())
}
