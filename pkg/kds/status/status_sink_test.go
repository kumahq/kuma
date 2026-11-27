package status_test

import (
	"bytes"
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/kds/status"
	kuma_log "github.com/kumahq/kuma/v3/pkg/log"
)

type staticAccessor struct{}

func (staticAccessor) GetStatus() (string, *system_proto.KDSSubscription) {
	return "zone-1", &system_proto.KDSSubscription{Id: "1"}
}

type errorStore struct {
	err error
}

func (s errorStore) Upsert(context.Context, string, *system_proto.KDSSubscription) error {
	return s.err
}

var _ = Describe("ZoneInsightSink", func() {
	flushOnStop := func(err error) string {
		logBuf := &bytes.Buffer{}
		ticker := func() *time.Ticker { return time.NewTicker(time.Hour) }
		sink := status.NewZoneInsightSink(
			staticAccessor{},
			ticker,
			ticker,
			time.Millisecond,
			errorStore{err: err},
			kuma_log.NewLoggerTo(logBuf, kuma_log.InfoLevel),
			context.Background(),
		)
		stop := make(chan struct{})
		close(stop)
		sink.Start(context.Background(), stop)
		return logBuf.String()
	}

	It("should not log an error when the owning Zone does not exist yet", func() {
		// when
		logs := flushOnStop(store.ErrorResourceNotFound(system.ZoneType, "zone-1", core_model.NoMesh))

		// then
		Expect(logs).ToNot(ContainSubstring("failed to flush zone status"))
	})

	It("should log an error on other failures", func() {
		// when
		logs := flushOnStop(errors.New("connection refused"))

		// then
		Expect(logs).To(ContainSubstring("failed to flush zone status"))
	})
})
