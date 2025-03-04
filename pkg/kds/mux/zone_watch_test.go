package mux_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

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

		rm = manager.NewResourceManager(memory.NewStore())
		zoneRes := system.NewZoneResource()
		zoneRes.Spec.Enabled = wrapperspb.Bool(true)
		Expect(rm.Create(
			context.Background(),
			zoneRes,
			store.CreateByKey(zone, core_model.NoMesh),
		)).To(Succeed())
		zoneInsight := system.NewZoneInsightResource()
		zoneInsight.Spec.HealthCheck = &system_proto.HealthCheck{
			Time: timestamppb.New(time.Now()),
		}
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
			switch event.(type) {
			case service.ZoneWentOffline:
				return true
			case service.StreamCancelled:
				return true
			default:
				return false
			}
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

	It("should timeout zones that stop sending a health check", func() {
		eventBus.Send(service.ZoneOpenedStream{
			TenantID: "",
			Zone:     zone,
		})

		// wait for opened stream to be registered
		// in real conditions the interval will be large enough
		// that these events will almost certainly be handled
		// by the ZoneWatch loop between polls and before the timeout
		time.Sleep(pollInterval)

		Eventually(timeouts.Recv(), zoneWentOfflineCheckTimeout).Should(Receive(Equal(service.ZoneWentOffline{
			TenantID: "",
			Zone:     zone,
		})))

		Consistently(timeouts.Recv(), zoneWentOfflineCheckTimeout).ShouldNot(Receive())
	})
	It("shouldn't timeout immediately if zoneinsight time is old", func() {
		zoneInsight := system.NewZoneInsightResource()
		Expect(rm.Get(
			context.Background(),
			zoneInsight,
			store.GetByKey(zone, core_model.NoMesh),
		)).To(Succeed())
		zoneInsight.Spec.HealthCheck = &system_proto.HealthCheck{
			Time: timestamppb.New(time.Now().AddDate(0, 0, -1)),
		}
		Expect(rm.Update(
			context.Background(),
			zoneInsight,
		)).To(Succeed())

		eventBus.Send(service.ZoneOpenedStream{
			TenantID: "",
			Zone:     zone,
		})

		// wait for opened stream to be registered
		// in real conditions the interval will be large enough
		// that these events will almost certainly be handled
		// by the ZoneWatch loop between polls and before the timeout
		time.Sleep(1 * pollInterval)

		Expect(timeouts.Recv()).NotTo(Receive())
	})
	It("shouldn't timeout as long as ZoneInsight is updated", func() {
		eventBus.Send(service.ZoneOpenedStream{
			TenantID: "",
			Zone:     zone,
		})

		// wait for opened stream to be registered
		// in real conditions the interval will be large enough
		// that these events will almost certainly be handled
		// by the ZoneWatch loop between polls and before the timeout
		time.Sleep(pollInterval)

		// Send a health check and block for a poll interval and make sure
		// nothing has been received
		// do this until we know the timeout would have come if weren't sending
		// health checks
		Consistently(func(g Gomega) {
			sendHealthCheckPing(rm, zone)
			g.Consistently(timeouts.Recv(), pollInterval).ShouldNot(Receive())
		}, 3*timeout).Should(Succeed())

		Eventually(timeouts.Recv(), zoneWentOfflineCheckTimeout).Should(Receive(Equal(service.ZoneWentOffline{
			TenantID: "",
			Zone:     zone,
		})))
	})
	It("should timeout if the zone is deleted", func() {
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
		Expect(rm.Delete(
			context.Background(),
			system.NewZoneInsightResource(),
			store.DeleteByKey(zone, core_model.NoMesh),
		)).To(Succeed())

		Eventually(timeouts.Recv(), zoneWentOfflineCheckTimeout).Should(Receive(Equal(service.ZoneWentOffline{
			TenantID: "",
			Zone:     zone,
		})))
	})
	It("should disconnect current stream when the same zone connects", func() {
		zoneInsight := system.NewZoneInsightResource()
		Expect(rm.Get(
			context.Background(),
			zoneInsight,
			store.GetByKey(zone, core_model.NoMesh),
		)).To(Succeed())
		zoneInsight.Spec.HealthCheck = &system_proto.HealthCheck{
			Time: timestamppb.New(time.Now()),
		}
		Expect(rm.Update(
			context.Background(),
			zoneInsight,
		)).To(Succeed())

		firstConnectTime := time.Now()
		eventBus.Send(service.ZoneOpenedStream{
			TenantID: "",
			Zone:     zone,
			Type:     service.GlobalToZone,
			ConnTime: firstConnectTime,
		})

		// wait for opened stream to be registered
		// in real conditions the interval will be large enough
		// that these events will almost certainly be handled
		// by the ZoneWatch loop between polls and before the timeout
		time.Sleep(1 * pollInterval)

		Expect(timeouts.Recv()).NotTo(Receive())

		// try to connect the same zone but on the 2nd stream
		eventBus.Send(service.ZoneOpenedStream{
			TenantID: "",
			Zone:     zone,
			Type:     service.GlobalToZone,
			ConnTime: time.Now(),
		})
		time.Sleep(1 * pollInterval)

		Eventually(timeouts.Recv(), 2*pollInterval).Should(Receive(Equal(service.StreamCancelled{
			TenantID: "",
			Zone:     zone,
			Type:     service.GlobalToZone,
			ConnTime: firstConnectTime,
		})))
	})
	It("should disconnect current stream when newer connection exists", func() {
		stopPing := make(chan struct{})
		oldConnection := time.Now()
		zoneInsight := system.NewZoneInsightResource()
		Expect(rm.Get(
			context.Background(),
			zoneInsight,
			store.GetByKey(zone, core_model.NoMesh),
		)).To(Succeed())
		zoneInsight.Spec.HealthCheck = &system_proto.HealthCheck{
			Time: timestamppb.New(time.Now()),
		}
		zoneInsight.Spec.KdsStreams = &system_proto.KDSStreams{
			GlobalToZone: &system_proto.KDSStream{
				GlobalInstanceId: "1",
				ConnectTime:      timestamppb.New(time.Now()),
			},
		}
		Expect(rm.Update(
			context.Background(),
			zoneInsight,
		)).To(Succeed())

		// Start a Goroutine for periodic health check pings
		go func() {
			for {
				select {
				case <-stopPing:
					return
				default:
					sendHealthCheckPing(rm, zone)
					time.Sleep(50 * time.Millisecond)
				}
			}
		}()

		// create a new connection which has time older than previous connection

		eventBus.Send(service.ZoneOpenedStream{
			TenantID: "",
			Zone:     zone,
			Type:     service.GlobalToZone,
			ConnTime: oldConnection,
		})

		// expect to cancel previous stream
		Eventually(timeouts.Recv(), zoneWentOfflineCheckTimeout).Should(Receive(Equal(service.StreamCancelled{
			TenantID: "",
			Zone:     zone,
			Type:     service.GlobalToZone,
			ConnTime: oldConnection,
		})))

		close(stopPing)
	})
})
