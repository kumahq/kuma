package events

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	kuma_events "github.com/kumahq/kuma/v3/pkg/events"
	kuma_v1alpha1 "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

type schemeManager struct {
	manager.Manager
	scheme *runtime.Scheme
}

func (m *schemeManager) GetScheme() *runtime.Scheme {
	return m.scheme
}

type recordingEmitter struct {
	events []kuma_events.Event
}

func (e *recordingEmitter) Send(event kuma_events.Event) {
	e.events = append(e.events, event)
}

func TestKubernetesObjectFromEvent(t *testing.T) {
	mesh := &kuma_v1alpha1.Mesh{
		ObjectMeta: metav1.ObjectMeta{Name: "mesh-1"},
	}

	t.Run("returns Kubernetes object directly", func(t *testing.T) {
		got, ok := kubernetesObjectFromEvent(mesh)
		if !ok {
			t.Fatalf("expected Kubernetes object")
		}
		if got != mesh {
			t.Fatalf("expected original object, got %#v", got)
		}
	})

	t.Run("unwraps DeletedFinalStateUnknown", func(t *testing.T) {
		got, ok := kubernetesObjectFromEvent(cache.DeletedFinalStateUnknown{Obj: mesh})
		if !ok {
			t.Fatalf("expected Kubernetes object from tombstone")
		}
		if got != mesh {
			t.Fatalf("expected unwrapped object, got %#v", got)
		}
	})

	t.Run("rejects unexpected payload", func(t *testing.T) {
		if got, ok := kubernetesObjectFromEvent(cache.DeletedFinalStateUnknown{Obj: "mesh-1"}); ok {
			t.Fatalf("expected tombstone with string payload to be rejected, got %#v", got)
		}
	})
}

func TestListenerSkipsUnexpectedObjects(t *testing.T) {
	droppedEvents := prometheus.NewCounter(prometheus.CounterOpts{Name: "test_dropped_events"})
	emitter := &recordingEmitter{}
	l := &listener{
		out:           emitter,
		droppedEvents: droppedEvents,
	}

	l.OnAdd("mesh-1", false)
	l.OnUpdate(nil, cache.DeletedFinalStateUnknown{Obj: "mesh-1"})
	l.OnDelete(cache.DeletedFinalStateUnknown{Obj: "mesh-1"})

	if len(emitter.events) != 0 {
		t.Fatalf("expected no events for unexpected payloads, got %#v", emitter.events)
	}
	if got := counterValue(t, droppedEvents); got != 3 {
		t.Fatalf("expected dropped event counter to be 3, got %f", got)
	}
}

func TestListenerHandlersAcceptDeletedFinalStateUnknown(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := kuma_v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add Kuma scheme: %v", err)
	}

	testCases := []struct {
		name      string
		operation kuma_events.Op
		call      func(*listener, any)
	}{
		{
			name:      "add",
			operation: kuma_events.Create,
			call: func(l *listener, obj any) {
				l.OnAdd(obj, false)
			},
		},
		{
			name:      "update",
			operation: kuma_events.Update,
			call: func(l *listener, obj any) {
				l.OnUpdate(nil, obj)
			},
		},
		{
			name:      "delete",
			operation: kuma_events.Delete,
			call: func(l *listener, obj any) {
				l.OnDelete(obj)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			emitter := &recordingEmitter{}
			l := &listener{
				mgr: &schemeManager{scheme: scheme},
				out: emitter,
			}

			tc.call(l, cache.DeletedFinalStateUnknown{
				Key: "mesh-1",
				Obj: &kuma_v1alpha1.Mesh{
					ObjectMeta: metav1.ObjectMeta{Name: "mesh-1"},
				},
			})

			if len(emitter.events) != 1 {
				t.Fatalf("expected one event, got %#v", emitter.events)
			}
			event, ok := emitter.events[0].(kuma_events.ResourceChangedEvent)
			if !ok {
				t.Fatalf("expected ResourceChangedEvent, got %T", emitter.events[0])
			}
			if event.Operation != tc.operation {
				t.Fatalf("expected %v operation, got %v", tc.operation, event.Operation)
			}
			if event.Type != core_model.ResourceType("Mesh") {
				t.Fatalf("expected Mesh type, got %q", event.Type)
			}
			if event.Key != (core_model.ResourceKey{Name: "mesh-1"}) {
				t.Fatalf("expected mesh-1 key, got %#v", event.Key)
			}
		})
	}
}

func counterValue(t *testing.T, counter prometheus.Counter) float64 {
	t.Helper()
	metric := &dto.Metric{}
	if err := counter.Write(metric); err != nil {
		t.Fatalf("failed to read counter: %v", err)
	}
	return metric.GetCounter().GetValue()
}
