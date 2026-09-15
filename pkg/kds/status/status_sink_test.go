package status_test

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	config_store "github.com/kumahq/kuma/v3/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/managers/apis/zoneinsight"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/kds/status"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
)

var _ = Describe("ZoneInsightSink", func() {
	It("should not log an ERROR when the owning Zone does not exist yet, only retry on the next tick", func() {
		// given a ZoneInsight store backed by the real manager, but the "zone-fed" Zone was not created yet
		// (reproduces the startup race between the "defaults" component creating the Zone and the KDS status
		// flusher trying to upsert its ZoneInsight)
		resManager := zoneinsight.NewZoneInsightManager(memory.NewStore(), &kuma_cp.ZoneMetrics{})
		zoneInsightStore := status.NewZonesInsightStore(resManager, config_store.UpsertConfig{}, false, core_store.NoTransactions{})

		recorded := &recordingLogSink{}
		log := logr.New(recorded)

		ticks := make(chan time.Time)
		sink := status.NewZoneInsightSink(
			&staticStatusAccessor{zone: "zone-fed", subscription: &system_proto.KDSSubscription{Id: "sub-1"}},
			func() *time.Ticker { return &time.Ticker{C: ticks} },
			func() *time.Ticker { return &time.Ticker{C: make(chan time.Time)} },
			time.Millisecond,
			zoneInsightStore,
			log,
			context.Background(),
		)

		stop := make(chan struct{})
		defer close(stop)
		go sink.Start(context.Background(), stop)

		// when the flusher runs before the owning Zone exists
		ticks <- time.Now()

		// then it retries quietly instead of logging an ERROR
		Eventually(func() []string {
			recorded.mu.Lock()
			defer recorded.mu.Unlock()
			return recorded.infoMsgs
		}, "1s", "5ms").Should(ContainElement("failed to flush ZoneInsight because the owning Zone does not exist yet. Will retry in the next tick"))

		recorded.mu.Lock()
		defer recorded.mu.Unlock()
		Expect(recorded.errorMsgs).To(BeEmpty())
	})
})

type staticStatusAccessor struct {
	zone         string
	subscription *system_proto.KDSSubscription
}

func (s *staticStatusAccessor) GetStatus() (string, *system_proto.KDSSubscription) {
	return s.zone, s.subscription
}

type recordingLogSink struct {
	mu        sync.Mutex
	infoMsgs  []string
	errorMsgs []string
}

func (r *recordingLogSink) Init(_ logr.RuntimeInfo) {}

func (r *recordingLogSink) Enabled(_ int) bool { return true }

func (r *recordingLogSink) Info(_ int, msg string, _ ...any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.infoMsgs = append(r.infoMsgs, msg)
}

func (r *recordingLogSink) Error(_ error, msg string, _ ...any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errorMsgs = append(r.errorMsgs, msg)
}

func (r *recordingLogSink) WithValues(_ ...any) logr.LogSink { return r }

func (r *recordingLogSink) WithName(_ string) logr.LogSink { return r }
