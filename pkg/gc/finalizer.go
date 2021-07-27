package gc

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

var (
	finalizerLog = core.Log.WithName("finalizer")
)

type finalizer struct {
	rm      manager.ResourceManager
	polling time.Duration
}

func NewFinalizer(rm manager.ResourceManager, polling time.Duration) component.Component {
	return &finalizer{
		rm:      rm,
		polling: polling,
	}
}

func (f *finalizer) Start(stop <-chan struct{}) error {
	ticker := time.NewTicker(f.polling)
	defer ticker.Stop()
	finalizerLog.Info("started")
	for {
		select {
		case <-ticker.C:
			if err := f.finalizeDataplaneInsights(); err != nil {
				finalizerLog.Error(err, "unable to finalize Dataplane Insights")
			}
			if err := f.finalizeZoneInsights(); err != nil {
				finalizerLog.Error(err, "unable to finalize Zone Insights")
			}
			if err := f.finalizeZoneIngressInsights(); err != nil {
				finalizerLog.Error(err, "unable to finalize Zone Ingress Insights")
			}
		case <-stop:
			finalizerLog.Info("stopped")
			return nil
		}
	}
}

func (f *finalizer) finalizeDataplaneInsights() error {
	ctx := context.Background()
	dataplaneInsights := &core_mesh.DataplaneInsightResourceList{}
	if err := f.rm.List(ctx, dataplaneInsights); err != nil {
		return err
	}
	for _, di := range dataplaneInsights.Items {
		if !di.Spec.IsOnline() {
			continue
		}
		if di.Spec.GetLastSubscription().GetCandidateForDisconnect() {
			finalizerLog.Info("mark data plane proxy as disconnected", "name", di.GetMeta().GetName(), "mesh", di.GetMeta().GetMesh())
			di.Spec.GetLastSubscription().DisconnectTime = timestamppb.New(core.Now())
		} else {
			finalizerLog.Info("mark data plane proxy as a candidate for disconnect", "name", di.GetMeta().GetName(), "mesh", di.GetMeta().GetMesh())
			di.Spec.GetLastSubscription().CandidateForDisconnect = true
		}
		err := manager.Upsert(f.rm, model.MetaToResourceKey(di.GetMeta()), core_mesh.NewDataplaneInsightResource(), func(resource model.Resource) bool {
			insight := resource.(*core_mesh.DataplaneInsightResource)
			insight.Spec.UpdateSubscription(di.Spec.GetLastSubscription())
			return true
		})
		if err != nil {
			finalizerLog.Error(err, "unable to finalize data plane insight", "name", di.GetMeta().GetName(), "mesh", di.GetMeta().GetMesh())
		}
	}
	return nil
}

func (f *finalizer) finalizeZoneInsights() error {
	ctx := context.Background()
	zoneInsights := &system.ZoneInsightResourceList{}
	if err := f.rm.List(ctx, zoneInsights); err != nil {
		return err
	}
	for _, zi := range zoneInsights.Items {
		if !zi.Spec.IsOnline() {
			continue
		}
		lastSubscription := zi.Spec.GetLastSubscription()
		if lastSubscription == nil {
			continue
		}
		if lastSubscription.GetCandidateForDisconnect() {
			finalizerLog.Info("mark zone as disconnected", "name", zi.GetMeta().GetName())
			lastSubscription.DisconnectTime = timestamppb.New(core.Now())
		} else {
			finalizerLog.Info("mark zone as a candidate for disconnect", "name", zi.GetMeta().GetName())
			lastSubscription.CandidateForDisconnect = true
		}
		err := manager.Upsert(f.rm, model.MetaToResourceKey(zi.GetMeta()), system.NewZoneInsightResource(), func(resource model.Resource) bool {
			insight := resource.(*system.ZoneInsightResource)
			insight.Spec.UpdateSubscription(lastSubscription)
			return true
		})
		if err != nil {
			finalizerLog.Error(err, "unable to finalize data plane insight", "name", zi.GetMeta().GetName(), "mesh", zi.GetMeta().GetMesh())
		}
	}
	return nil
}

func (f *finalizer) finalizeZoneIngressInsights() error {
	ctx := context.Background()
	zoneIngressInsights := &core_mesh.ZoneIngressInsightResourceList{}
	if err := f.rm.List(ctx, zoneIngressInsights); err != nil {
		return err
	}
	for _, zi := range zoneIngressInsights.Items {
		if !zi.Spec.IsOnline() {
			continue
		}
		if zi.Spec.GetLastSubscription().GetCandidateForDisconnect() {
			finalizerLog.Info("mark zone ingress as disconnected", "name", zi.GetMeta().GetName(), "mesh", zi.GetMeta().GetMesh())
			zi.Spec.GetLastSubscription().DisconnectTime = timestamppb.New(core.Now())
		} else {
			finalizerLog.Info("mark zone ingress as a candidate for disconnect", "name", zi.GetMeta().GetName(), "mesh", zi.GetMeta().GetMesh())
			zi.Spec.GetLastSubscription().CandidateForDisconnect = true
		}
		err := manager.Upsert(f.rm, model.MetaToResourceKey(zi.GetMeta()), core_mesh.NewZoneIngressInsightResource(), func(resource model.Resource) bool {
			insight := resource.(*core_mesh.ZoneIngressInsightResource)
			insight.Spec.UpdateSubscription(zi.Spec.GetLastSubscription())
			return true
		})
		if err != nil {
			finalizerLog.Error(err, "unable to finalize data plane insight", "name", zi.GetMeta().GetName(), "mesh", zi.GetMeta().GetMesh())
		}
	}
	return nil
}

func (f *finalizer) NeedLeaderElection() bool {
	return true
}
