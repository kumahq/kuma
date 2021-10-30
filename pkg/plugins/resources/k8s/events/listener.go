package events

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/events"
	kuma_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

var log = core.Log.WithName("k8s-event-listener")

type listener struct {
	mgr manager.Manager
	out events.Emitter
}

func NewListener(mgr manager.Manager, out events.Emitter) component.Component {
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
		go func(typ core_model.ResourceType) {
			log.V(1).Info("start watching resource", "type", typ)
			informer.Run(stop)
		}(t)
	}
	return nil
}

func resourceKey(obj model.KubernetesObject) core_model.ResourceKey {
	var name string
	if obj.Scope() == model.ScopeCluster {
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
	if err := k.addTypeInformationToObject(kobj); err != nil {
		log.Error(err, "unable to add TypeMeta to KubernetesObject")
		return
	}
	k.out.Send(events.ResourceChangedEvent{
		Operation: events.Create,
		Type:      core_model.ResourceType(kobj.GetObjectKind().GroupVersionKind().Kind),
		Key:       resourceKey(kobj),
	})
}

func (k *listener) OnUpdate(oldObj, newObj interface{}) {
	kobj := newObj.(model.KubernetesObject)
	if err := k.addTypeInformationToObject(kobj); err != nil {
		log.Error(err, "unable to add TypeMeta to KubernetesObject")
		return
	}
	k.out.Send(events.ResourceChangedEvent{
		Operation: events.Update,
		Type:      core_model.ResourceType(kobj.GetObjectKind().GroupVersionKind().Kind),
		Key:       resourceKey(kobj),
	})
}

func (k *listener) OnDelete(obj interface{}) {
	kobj := obj.(model.KubernetesObject)
	if err := k.addTypeInformationToObject(kobj); err != nil {
		log.Error(err, "unable to add TypeMeta to KubernetesObject")
		return
	}
	k.out.Send(events.ResourceChangedEvent{
		Operation: events.Delete,
		Type:      core_model.ResourceType(kobj.GetObjectKind().GroupVersionKind().Kind),
		Key:       resourceKey(kobj),
	})
}

func (k *listener) NeedLeaderElection() bool {
	return false
}

func (k *listener) addTypeInformationToObject(obj runtime.Object) error {
	gvks, _, err := k.mgr.GetScheme().ObjectKinds(obj)
	if err != nil {
		return errors.Wrap(err, "missing apiVersion or kind and cannot assign it")
	}

	for _, gvk := range gvks {
		if len(gvk.Kind) == 0 {
			continue
		}
		if len(gvk.Version) == 0 || gvk.Version == runtime.APIVersionInternal {
			continue
		}
		obj.GetObjectKind().SetGroupVersionKind(gvk)
		break
	}

	return nil
}

func (k *listener) createListerWatcher(gvk schema.GroupVersionKind) (cache.ListerWatcher, error) {
	mapping, err := k.mgr.GetRESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	client, err := apiutil.RESTClientForGVK(gvk, false, k.mgr.GetConfig(), serializer.NewCodecFactory(k.mgr.GetScheme()))
	if err != nil {
		return nil, err
	}
	listGVK := gvk.GroupVersion().WithKind(gvk.Kind + "List")
	listObj, err := k.mgr.GetScheme().New(listGVK)
	if err != nil {
		return nil, err
	}
	paramCodec := runtime.NewParameterCodec(k.mgr.GetScheme())
	ctx := context.Background()
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
