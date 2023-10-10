package mux_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/timestamppb"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/multizone"
	"github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/kds/mux"
	"github.com/kumahq/kuma/pkg/kds/service"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

func sendHealthCheckPing(rm manager.ResourceManager, name string) {
	zoneInsight := system.NewZoneInsightResource()
	Expect(rm.Get(
		context.Background(),
		zoneInsight,
		store.GetByKey(name, core_model.NoMesh),
	)).To(Succeed())

	zoneInsight.Spec.HealthCheck = &system_proto.HealthCheck{
		Time: timestamppb.New(time.Now()),
	}
	Expect(rm.Update(
		context.Background(),
		zoneInsight,
	)).To(Succeed())
}

const zone = "zone-1"

var _ = Describe("ZoneWatch", func() {
	var errCh chan error

	var rm manager.ResourceManager
	var eventBus events.EventBus
	var stop chan struct{}
	var zoneWatch *mux.ZoneWatch
	var timeouts events.Listener

	pollInterval := 100 * time.Millisecond
	timeout := 5 * pollInterval

	BeforeEach(func() {
		metrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		eventBus, err = events.NewEventBus(10, metrics)
		Expect(err).NotTo(HaveOccurred())

		cfg := multizone.ZoneHealthCheckConfig{
			PollInterval: types.Duration{Duration: pollInterval},
			Timeout:      types.Duration{Duration: timeout},
		}

		zoneInsight := system.NewZoneInsightResource()
		zoneInsight.Spec.HealthCheck = &system_proto.HealthCheck{
			Time: timestamppb.New(time.Now()),
		}
		rm = manager.NewResourceManager(memory.NewStore())
		Expect(rm.Create(
			context.Background(),
			zoneInsight,
			store.CreateByKey(zone, core_model.NoMesh),
		)).To(Succeed())

		log := core.Log.WithName("test")
		zoneWatch, err = mux.NewZoneWatch(
			log,
			cfg,
			metrics,
			eventBus,
			rm,
			context.Background(),
		)
		Expect(err).NotTo(HaveOccurred())

		stop = make(chan struct{})

		timeouts = eventBus.Subscribe(func(event events.Event) bool {
			_, ok := event.(service.ZoneWentOffline)
			return ok
		})

		errCh = make(chan error, 1)

		go func() {
			errCh <- zoneWatch.Start(stop)
		}()

		// wait for ZoneWatch to have subscribed to new zone events
		time.Sleep(pollInterval)
	})

	AfterEach(func() {
		select {
		case <-errCh:
			Fail("zone watch should not have stopped")
		default:
		}
		close(stop)
		timeouts.Close()
		Eventually(errCh).Should(Receive(Succeed()))
	})

	// We know _best case_ the zone will register offline
	// in timeout + pollInterval
	zoneWentOfflineCheckTimeout := timeout + 2*pollInterval

	Describe("basic", FlakeAttempts(3), func() {
		It("should timeout zones that stop sending a health check", func() {
			sendHealthCheckPing(rm, zone)
			eventBus.Send(service.ZoneOpenedStream{
				TenantID: "",
				Zone:     zone,
			})

			// wait for opened stream to be registered
			// in real conditions the interval will be large enough
			// that these events will almost certainly be handled
			// by the ZoneWatch loop between polls and before the timeout
			time.Sleep(pollInterval)

			sendHealthCheckPing(rm, zone)

			Eventually(timeouts.Recv(), zoneWentOfflineCheckTimeout).Should(Receive(Equal(service.ZoneWentOffline{
				TenantID: "",
				Zone:     zone,
			})))
		})
		It("should not timeout zones that haven't sent a health check yet", func() {
			sendHealthCheckPing(rm, zone)
			eventBus.Send(service.ZoneOpenedStream{
				TenantID: "",
				Zone:     zone,
			})

			Consistently(timeouts.Recv(), zoneWentOfflineCheckTimeout).ShouldNot(Receive())
		})
	})
})
