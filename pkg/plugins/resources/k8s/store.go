package k8s

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kumahq/kuma/pkg/core/resources/registry"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	common_k8s "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	util_k8s "github.com/kumahq/kuma/pkg/util/k8s"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ store.ResourceStore = &KubernetesStore{}

type KubernetesStore struct {
	Client    kube_client.Client
	Converter Converter
	Scheme    *kube_runtime.Scheme
}

func NewStore(client kube_client.Client, scheme *kube_runtime.Scheme) (store.ResourceStore, error) {
	return &KubernetesStore{
		Client:    client,
		Converter: DefaultConverter(),
		Scheme:    scheme,
	}, nil
}

func (s *KubernetesStore) Create(ctx context.Context, r core_model.Resource, fs ...store.CreateOptionsFunc) error {
	opts := store.NewCreateOptions(fs...)
	obj, err := s.Converter.ToKubernetesObject(r)
	if err != nil {
		return errors.Wrap(err, "failed to convert core model into k8s counterpart")
	}
	name, namespace, err := k8sNameNamespace(opts.Name, obj.Scope())
	if err != nil {
		return err
	}
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

	if opts.Synced {
		markAsSynced(obj)
	}

	if err := s.Client.Create(ctx, obj); err != nil {
		if kube_apierrs.IsAlreadyExists(err) {
			return store.ErrorResourceAlreadyExists(r.GetType(), opts.Name, opts.Mesh)
		}
		return errors.Wrap(err, "failed to create k8s resource")
	}
	err = s.Converter.ToCoreResource(obj, r)
	if err != nil {
		return errors.Wrap(err, "failed to convert k8s model into core counterpart")
	}
	return nil
}

func markAsSynced(obj k8s_model.KubernetesObject) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[common_k8s.K8sSynced] = "true"
	obj.SetAnnotations(annotations)
}

func (s *KubernetesStore) Update(ctx context.Context, r core_model.Resource, fs ...store.UpdateOptionsFunc) error {
	opts := store.NewUpdateOptions(fs...)

	obj, err := s.Converter.ToKubernetesObject(r)
	if err != nil {
		return errors.Wrapf(err, "failed to convert core model of type %s into k8s counterpart", r.GetType())
	}

	if opts.Synced {
		markAsSynced(obj)
	}

	if err := s.Client.Update(ctx, obj); err != nil {
		if kube_apierrs.IsConflict(err) {
			return store.ErrorResourceConflict(r.GetType(), r.GetMeta().GetName(), r.GetMeta().GetMesh())
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
		return errors.Wrapf(err, "failed to convert core model of type %s into k8s counterpart", r.GetType())
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
		return errors.Wrapf(err, "failed to convert core model of type %s into k8s counterpart", r.GetType())
	}
	name, namespace, err := k8sNameNamespace(opts.Name, obj.Scope())
	if err != nil {
		return err
	}
	if err := s.Client.Get(ctx, kube_client.ObjectKey{Namespace: namespace, Name: name}, obj); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return store.ErrorResourceNotFound(r.GetType(), opts.Name, opts.Mesh)
		}
		return errors.Wrap(err, "failed to get k8s resource")
	}
	if err := s.Converter.ToCoreResource(obj, r); err != nil {
		return errors.Wrap(err, "failed to convert k8s model into core counterpart")
	}
	if opts.Version != "" && r.GetMeta().GetVersion() != opts.Version {
		return store.ErrorResourcePreconditionFailed(r.GetType(), opts.Name, opts.Mesh)
	}
	if r.GetMeta().GetMesh() != opts.Mesh {
		return store.ErrorResourceNotFound(r.GetType(), opts.Name, opts.Mesh)
	}
	return nil
}
func (s *KubernetesStore) List(ctx context.Context, rs core_model.ResourceList, fs ...store.ListOptionsFunc) error {
	opts := store.NewListOptions(fs...)
	obj, err := s.Converter.ToKubernetesList(rs)
	if err != nil {
		return errors.Wrapf(err, "failed to convert core list model of type %s into k8s counterpart", rs.GetItemType())
	}
	if err := s.Client.List(ctx, obj); err != nil {
		return errors.Wrap(err, "failed to list k8s resources")
	}
	predicate := func(r core_model.Resource) bool {
		if opts.Mesh != "" {
			return r.GetMeta().GetMesh() == opts.Mesh
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
	total := len(fullList.GetItems())
	if opts.PageSize > 0 {
		offset := 0
		if opts.PageOffset != "" {
			if o, err := strconv.Atoi(opts.PageOffset); err != nil {
				return store.ErrorInvalidOffset
			} else {
				offset = o
			}
		}

		min := func(x, y int) int {
			if x < y {
				return x
			} else {
				return y
			}
		}
		for _, item := range fullList.GetItems()[offset:min(offset+opts.PageSize, total)] {
			_ = rs.AddItem(item)
		}
		nextOffset := ""
		if offset+opts.PageSize < total { // set new offset only if we did not reach the end of the collection
			nextOffset = strconv.Itoa(offset + opts.PageSize)
		}
		rs.GetPagination().SetNextOffset(nextOffset)
	} else {
		for _, item := range fullList.GetItems() {
			_ = rs.AddItem(item)
		}
	}
	rs.GetPagination().SetTotal(uint32(total))
	return nil
}

func k8sNameNamespace(coreName string, scope k8s_model.Scope) (string, string, error) {
	switch scope {
	case k8s_model.ScopeCluster:
		return coreName, "", nil
	case k8s_model.ScopeNamespace:
		return util_k8s.CoreNameToK8sName(coreName)
	default:
		return "", "", errors.Errorf("unknown scope %s", scope)
	}
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
	return common_k8s.ResourceNameExtensions(m.ObjectMeta.Namespace, m.ObjectMeta.Name)
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

type ConverterPredicate = func(core_model.Resource) bool
type Converter interface {
	ToKubernetesObject(core_model.Resource) (k8s_model.KubernetesObject, error)
	ToKubernetesList(core_model.ResourceList) (k8s_model.KubernetesList, error)
	ToCoreResource(obj k8s_model.KubernetesObject, out core_model.Resource) error
	ToCoreList(obj k8s_model.KubernetesList, out core_model.ResourceList, predicate ConverterPredicate) error
}

var once sync.Once
var sc Converter

const expirationTime = 5 * time.Minute

func DefaultConverter() Converter {
	once.Do(func() {
		sc = &SimpleConverter{
			KubeFactory: &SimpleKubeFactory{
				KubeTypes: k8s_registry.Global(),
			},
			Cache: cache.New(expirationTime, time.Duration(int64(float64(expirationTime)*0.9))),
		}
	})
	return sc
}

var _ Converter = &SimpleConverter{}

type SimpleConverter struct {
	KubeFactory KubeFactory
	Cache       *cache.Cache
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
	out.SetMeta(&KubernetesMetaAdapter{*obj.GetObjectMeta(), obj.GetMesh()})
	key := strings.Join([]string{
		obj.GetNamespace(),
		obj.GetName(),
		obj.GetMesh(),
		obj.GetResourceVersion(),
		proto.MessageName(out.GetSpec()),
	}, ":")
	if v, ok := c.Cache.Get(key); ok {
		return out.SetSpec(v.(core_model.ResourceSpec))
	}
	if err := util_proto.FromMap(obj.GetSpec(), out.GetSpec()); err != nil {
		return err
	}
	c.Cache.SetDefault(key, out.GetSpec())
	return nil
}

func (c *SimpleConverter) ToCoreList(in k8s_model.KubernetesList, out core_model.ResourceList, predicate ConverterPredicate) error {
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
