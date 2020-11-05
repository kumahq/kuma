package events

import (
	"context"
	"fmt"

	kuma_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

type listener struct {
	mgr manager.Manager
	out events.Writer
}

func NewListener(mgr manager.Manager, out events.Writer) component.Component {
	return &listener{
		mgr: mgr,
		out: out,
	}
}

func (k *listener) Start(stop <-chan struct{}) error {
	types := core_registry.Global().ObjectTypes()
	knownTypes := k.mgr.GetScheme().KnownTypes(kuma_v1alpha1.GroupVersion)
	for _, t := range types {
		if _, ok := knownTypes[string(t)]; !ok {
			continue
		}
		gvk := kuma_v1alpha1.GroupVersion.WithKind(string(t))
		lw, err := k.createListerWatcher(gvk)
		if err != nil {
			return err
		}
		coreObj, err := core_registry.Global().NewObject(t)
		if err != nil {
			return err
		}
		obj, err := k8s_registry.Global().NewObject(coreObj.GetSpec())
		if err != nil {
			return err
		}
		informer := cache.NewSharedInformer(lw, obj, 0)
		informer.AddEventHandler(k)
		go func() {
			informer.Run(stop)
		}()
	}
	return nil
}

func resourceKey(obj model.KubernetesObject) core_model.ResourceKey {
	var name string
	if obj.GetNamespace() == "" {
		name = obj.GetName()
	} else {
		name = fmt.Sprintf("%s.%s", obj.GetName(), obj.GetNamespace())
	}
	return core_model.ResourceKey{
		Name: name,
		Mesh: obj.GetMesh(),
	}
}

func (k *listener) OnAdd(obj interface{}) {
	kobj := obj.(model.KubernetesObject)
	k.out.Send(
		events.Create,
		core_model.ResourceType(kobj.GetObjectKind().GroupVersionKind().Kind),
		resourceKey(kobj),
	)
}

func (k *listener) OnUpdate(oldObj, newObj interface{}) {
	kobj := newObj.(model.KubernetesObject)
	k.out.Send(
		events.Update,
		core_model.ResourceType(kobj.GetObjectKind().GroupVersionKind().Kind),
		resourceKey(kobj),
	)
}

func (k *listener) OnDelete(obj interface{}) {
	kobj := obj.(model.KubernetesObject)
	k.out.Send(
		events.Delete,
		core_model.ResourceType(kobj.GetObjectKind().GroupVersionKind().Kind),
		resourceKey(kobj),
	)
}

func (k *listener) NeedLeaderElection() bool {
	return false
}

func (k *listener) createListerWatcher(gvk schema.GroupVersionKind) (cache.ListerWatcher, error) {
	mapping, err := k.mgr.GetRESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	client, err := apiutil.RESTClientForGVK(gvk, k.mgr.GetConfig(), serializer.NewCodecFactory(k.mgr.GetScheme()))
	if err != nil {
		return nil, err
	}
	listGVK := gvk.GroupVersion().WithKind(gvk.Kind + "List")
	listObj, err := k.mgr.GetScheme().New(listGVK)
	if err != nil {
		return nil, err
	}
	paramCodec := runtime.NewParameterCodec(k.mgr.GetScheme())
	ctx := context.TODO()
	return &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			res := listObj.DeepCopyObject()
			err := client.Get().
				Resource(mapping.Resource.Resource).
				VersionedParams(&opts, paramCodec).
				Do(ctx).
				Into(res)
			return res, err
		},
		// Setup the watch function
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			// Watch needs to be set to true separately
			opts.Watch = true
			return client.Get().
				Resource(mapping.Resource.Resource).
				VersionedParams(&opts, paramCodec).
				Watch(ctx)
		},
	}, nil
}
