package status_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	config_store "github.com/kumahq/kuma/v3/pkg/config/core/resources/store"
	config_types "github.com/kumahq/kuma/v3/pkg/config/types"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/kds/status"
	memory_resources "github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

type conflictingManager struct {
	manager.ResourceManager
	conflicts int
}

func (m *conflictingManager) Update(ctx context.Context, r core_model.Resource, fs ...store.UpdateOptionsFunc) error {
	if m.conflicts > 0 {
		m.conflicts--
		return store.ErrorResourceConflict(r.Descriptor().Name, r.GetMeta().GetName(), r.GetMeta().GetMesh())
	}
	return m.ResourceManager.Update(ctx, r, fs...)
}

var _ = Describe("ZoneInsightStore", func() {
	It("should retry upsert on conflict so the disconnect is persisted", func() {
		// given
		resManager := &conflictingManager{ResourceManager: manager.NewResourceManager(memory_resources.NewStore())}
		insightStore := status.NewZonesInsightStore(
			resManager,
			config_store.UpsertConfig{
				ConflictRetryBaseBackoff: config_types.Duration{Duration: time.Millisecond},
				ConflictRetryMaxTimes:    5,
			},
			false,
			&store.NoTransactions{},
		)
		subscription := &system_proto.KDSSubscription{
			Id:          "sub-1",
			ConnectTime: util_proto.MustTimestampProto(time.Now()),
			Status:      system_proto.NewSubscriptionStatus(time.Now()),
		}
		Expect(insightStore.Upsert(context.Background(), "zone-1", subscription)).To(Succeed())

		// when another writer updates the insight concurrently with the final flush
		resManager.conflicts = 2
		subscription.DisconnectTime = util_proto.MustTimestampProto(time.Now())
		err := insightStore.Upsert(context.Background(), "zone-1", subscription)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resManager.conflicts).To(BeZero())
		insight := system.NewZoneInsightResource()
		Expect(resManager.Get(context.Background(), insight, store.GetByKey("zone-1", core_model.NoMesh))).To(Succeed())
		Expect(insight.Spec.IsOnline()).To(BeFalse())
	})
})
