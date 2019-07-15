package server

import (
	"context"
	"time"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/gogo/protobuf/proto"
)

type DataplaneStatusSink interface {
	Start(accessor DataplaneStatusAccessor, stop <-chan struct{})
}

type DataplaneStatusAccessor interface {
	GetStatusSnapshot() (core_model.ResourceKey, *mesh_proto.DiscoverySubscription)
}

func DefaultDataplaneStatusSink(rs core_store.ResourceStore) DataplaneStatusSink {
	return &dataplaneStatusSink{
		rs: rs,
	}
}

var _ DataplaneStatusSink = &dataplaneStatusSink{}

type dataplaneStatusSink struct {
	rs core_store.ResourceStore
}

func (f *dataplaneStatusSink) Start(accessor DataplaneStatusAccessor, stop <-chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var lastFlushedSnapshot *mesh_proto.DiscoverySubscription

	takeSnapshotAndFlush := func() {
		dataplaneId, snapshot := accessor.GetStatusSnapshot()
		if snapshot.Equal(lastFlushedSnapshot) {
			return
		}
		copy := proto.Clone(snapshot).(*mesh_proto.DiscoverySubscription)
		if err := f.flush(dataplaneId, copy); err != nil {
			xdsServerLog.Error(err, "failed to flush Dataplane status", "dataplaneid", dataplaneId)
		} else {
			lastFlushedSnapshot = snapshot
		}
	}

	for {
		select {
		case <-stop:
			takeSnapshotAndFlush()
			return
		case <-ticker.C:
			takeSnapshotAndFlush()
		}
	}
}

func (f *dataplaneStatusSink) flush(dataplaneId core_model.ResourceKey, subscription *mesh_proto.DiscoverySubscription) error {
	create := false
	dataplaneStatus := &mesh_core.DataplaneStatusResource{}
	err := f.rs.Get(context.Background(), dataplaneStatus, core_store.GetBy(dataplaneId))
	if err != nil {
		if core_store.IsResourceNotFound(err) {
			create = true
		} else {
			return err
		}
	}
	originalStatus := proto.Clone(&dataplaneStatus.Spec).(*mesh_proto.DataplaneStatus)
	dataplaneStatus.Spec.UpdateSubscription(subscription)
	if dataplaneStatus.Spec.Equal(originalStatus) {
		return nil
	}
	xdsServerLog.V(1).Info("going ahead with Dataplane status update", "dataplaneid", dataplaneId, "subscription", subscription)
	if create {
		return f.rs.Create(context.Background(), dataplaneStatus, core_store.CreateBy(dataplaneId))
	} else {
		return f.rs.Update(context.Background(), dataplaneStatus)
	}
}
