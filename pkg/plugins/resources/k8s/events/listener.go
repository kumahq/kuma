package events

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kumahq/kuma/v3/pkg/core"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_registry "github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v3/pkg/events"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	kuma_v1alpha1 "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
)

var log = core.Log.WithName("k8s-event-listener")

type listener struct {
	mgr           manager.Manager
	out           events.Emitter
	droppedEvents *prometheus.CounterVec
}

func NewListener(mgr manager.Manager, out events.Emitter, metrics core_metrics.Metrics) (component.Component, error) {
	droppedEvents := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "k8s_events_dropped_total",
		Help: "Number of Kubernetes informer events dropped because their payload could not be converted to a Kuma Kubernetes object",
	}, []string{"operation"})
	if err := metrics.Register(droppedEvents); err != nil {
		return nil, err
	}
	return &listener{
		mgr:           mgr,
		out:           out,
		droppedEvents: droppedEvents,
	}, nil
}

// Start registers the listener on the manager cache's informers, the same ones the
// Kubernetes store reads from, so every type is watched and held in memory once.
func (k *listener) Start(stop <-chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-stop:
		case <-ctx.Done():
		}
		cancel()
	}()

	types := core_registry.Global().ObjectTypes()
	knownTypes := k.mgr.GetScheme().KnownTypes(kuma_v1alpha1.GroupVersion)
	for _, t := range types {
		if _, ok := knownTypes[string(t)]; !ok {
			continue
		}
		coreObj, err := core_registry.Global().NewObject(t)
		if err != nil {
			return err
		}
		obj, err := k8s_registry.Global().NewObject(coreObj.GetSpec())
		if err != nil {
			return err
		}

		informer, err := k.mgr.GetCache().GetInformer(ctx, obj)
		if err != nil {
			return errors.Wrapf(err, "could not get informer for %s", t)
		}
		if _, err := informer.AddEventHandler(k); err != nil {
			return err
		}
		log.V(1).Info("start watching resource", "type", t)
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

func (k *listener) OnAdd(obj any, _ bool) {
	kobj, ok := kubernetesObjectFromEvent(obj)
	if !ok {
		k.recordDroppedEvent("add", obj)
		return
	}
	kind, err := k.kindOf(kobj)
	if err != nil {
		log.Error(err, "unable to resolve kind of KubernetesObject")
		return
	}
	k.out.Send(events.ResourceChangedEvent{
		Operation: events.Create,
		Type:      core_model.ResourceType(kind),
		Key:       resourceKey(kobj),
	})
}

func (k *listener) OnUpdate(oldObj, newObj any) {
	// Delete tombstones are expected on delete callbacks, but unwrap defensively
	// here too so malformed update payloads fail closed instead of panicking.
	kobj, ok := kubernetesObjectFromEvent(newObj)
	if !ok {
		k.recordDroppedEvent("update", newObj)
		return
	}
	// The manager cache resyncs periodically and replays unchanged objects as updates.
	if oldKobj, ok := kubernetesObjectFromEvent(oldObj); ok && oldKobj.GetResourceVersion() == kobj.GetResourceVersion() {
		return
	}
	kind, err := k.kindOf(kobj)
	if err != nil {
		log.Error(err, "unable to resolve kind of KubernetesObject")
		return
	}
	k.out.Send(events.ResourceChangedEvent{
		Operation: events.Update,
		Type:      core_model.ResourceType(kind),
		Key:       resourceKey(kobj),
	})
}

func (k *listener) OnDelete(obj any) {
	kobj, ok := kubernetesObjectFromEvent(obj)
	if !ok {
		k.recordDroppedEvent("delete", obj)
		return
	}
	kind, err := k.kindOf(kobj)
	if err != nil {
		log.Error(err, "unable to resolve kind of KubernetesObject")
		return
	}
	k.out.Send(events.ResourceChangedEvent{
		Operation: events.Delete,
		Type:      core_model.ResourceType(kind),
		Key:       resourceKey(kobj),
	})
}

func kubernetesObjectFromEvent(obj any) (model.KubernetesObject, bool) {
	if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		// Client-go stores the last observed object in the tombstone. If that
		// payload is itself another tombstone, treat it as malformed.
		obj = tombstone.Obj
	}

	kobj, ok := obj.(model.KubernetesObject)
	return kobj, ok
}

func (k *listener) recordDroppedEvent(operation string, obj any) {
	if k.droppedEvents != nil {
		k.droppedEvents.WithLabelValues(operation).Inc()
	}
	log.Error(errors.Errorf("unexpected object type on %s", operation), "skipping Kubernetes informer event", "type", fmt.Sprintf("%T", obj))
}

func (k *listener) NeedLeaderElection() bool {
	return false
}

// kindOf reads the Kind from the scheme instead of setting TypeMeta on the object:
// objects come from the shared manager cache and must not be mutated.
func (k *listener) kindOf(obj runtime.Object) (string, error) {
	gvks, _, err := k.mgr.GetScheme().ObjectKinds(obj)
	if err != nil {
		return "", errors.Wrap(err, "missing apiVersion or kind")
	}
	for _, gvk := range gvks {
		if gvk.Kind == "" || gvk.Version == "" || gvk.Version == runtime.APIVersionInternal {
			continue
		}
		return gvk.Kind, nil
	}
	return "", errors.Errorf("no versioned kind registered for %T", obj)
}
