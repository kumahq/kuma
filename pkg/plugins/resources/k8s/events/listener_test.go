package events

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	kuma_events "github.com/kumahq/kuma/v2/pkg/events"
	kuma_v1alpha1 "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s/native/api/v1alpha1"
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

func counterValue(counter prometheus.Counter) float64 {
	metric := &dto.Metric{}
	ExpectWithOffset(1, counter.Write(metric)).To(Succeed())
	return metric.GetCounter().GetValue()
}

var _ = Describe("kubernetesObjectFromEvent", func() {
	mesh := &kuma_v1alpha1.Mesh{
		ObjectMeta: metav1.ObjectMeta{Name: "mesh-1"},
	}

	It("returns Kubernetes object directly", func() {
		got, ok := kubernetesObjectFromEvent(mesh)
		Expect(ok).To(BeTrue())
		Expect(got).To(BeIdenticalTo(mesh))
	})

	It("unwraps DeletedFinalStateUnknown", func() {
		got, ok := kubernetesObjectFromEvent(cache.DeletedFinalStateUnknown{Obj: mesh})
		Expect(ok).To(BeTrue())
		Expect(got).To(BeIdenticalTo(mesh))
	})

	It("rejects unexpected payload", func() {
		_, ok := kubernetesObjectFromEvent(cache.DeletedFinalStateUnknown{Obj: "mesh-1"})
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("listener", func() {
	It("skips unexpected objects and records dropped events per operation", func() {
		droppedEvents := prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "test_dropped_events",
			Help: "Test counter for dropped Kubernetes events.",
		}, []string{"operation"})
		emitter := &recordingEmitter{}
		l := &listener{
			out:           emitter,
			droppedEvents: droppedEvents,
		}

		l.OnAdd("mesh-1", false)
		l.OnUpdate(nil, cache.DeletedFinalStateUnknown{Obj: "mesh-1"})
		l.OnDelete(cache.DeletedFinalStateUnknown{Obj: "mesh-1"})

		Expect(counterValue(droppedEvents.WithLabelValues("add"))).To(Equal(float64(1)))
		Expect(counterValue(droppedEvents.WithLabelValues("update"))).To(Equal(float64(1)))
		Expect(counterValue(droppedEvents.WithLabelValues("delete"))).To(Equal(float64(1)))
		Expect(emitter.events).To(BeEmpty())
	})

	DescribeTable("handlers accept DeletedFinalStateUnknown tombstones",
		func(operation kuma_events.Op, call func(*listener, any)) {
			scheme := runtime.NewScheme()
			Expect(kuma_v1alpha1.AddToScheme(scheme)).To(Succeed())

			emitter := &recordingEmitter{}
			l := &listener{
				mgr: &schemeManager{scheme: scheme},
				out: emitter,
			}

			call(l, cache.DeletedFinalStateUnknown{
				Key: "mesh-1",
				Obj: &kuma_v1alpha1.Mesh{
					ObjectMeta: metav1.ObjectMeta{Name: "mesh-1"},
				},
			})

			Expect(emitter.events).To(HaveLen(1))
			event, ok := emitter.events[0].(kuma_events.ResourceChangedEvent)
			Expect(ok).To(BeTrue())
			Expect(event.Operation).To(Equal(operation))
			Expect(event.Type).To(Equal(core_model.ResourceType("Mesh")))
			Expect(event.Key).To(Equal(core_model.ResourceKey{Name: "mesh-1"}))
		},
		Entry("add", kuma_events.Create, func(l *listener, obj any) {
			l.OnAdd(obj, false)
		}),
		Entry("update", kuma_events.Update, func(l *listener, obj any) {
			l.OnUpdate(nil, obj)
		}),
		Entry("delete", kuma_events.Delete, func(l *listener, obj any) {
			l.OnDelete(obj)
		}),
	)
})
