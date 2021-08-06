package gc

import (
	"context"
	"time"

	"github.com/pkg/errors"

	kuma_interfaces "github.com/kumahq/kuma/api/helpers"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

var (
	finalizerLog = core.Log.WithName("finalizer")
)

// subscriptionFinalizer is a component that allows to finalize subscriptions for Insights.
// The component iterates over all online Insights on every tick and marks their last
// subscriptions as a candidate for disconnection. If subscription has already been marked as
// a candidate then component finalizes the subscription by setting DisconnectTime.
//
// Every Insight has a statusSink that periodically updates Subscriptions. That's why if there
// is an active connection between Kuma CP and DPP (or Global CP and Zone CP) then Subscription
// will be unmarked back from the candidate status. If there is no active connection then after 2 ticks
// Insight will be considered Offline.
//
// This component allows to solve the corner case we had with Insights:
// 1. Kuma CP is down
// 2. DPP is down
// 3. Kuma CP is up
// 4. DPP status is Online whereas it should be Offline
type subscriptionFinalizer struct {
	rm        manager.ResourceManager
	newTicker func() *time.Ticker
	types     []core_model.ResourceType
}

func NewSubscriptionFinalizer(rm manager.ResourceManager, newTicker func() *time.Ticker, types ...core_model.ResourceType) component.Component {
	return &subscriptionFinalizer{
		rm:        rm,
		types:     types,
		newTicker: newTicker,
	}
}

func (f *subscriptionFinalizer) Start(stop <-chan struct{}) error {
	ticker := f.newTicker()
	defer ticker.Stop()
	for _, typ := range f.types {
		if err := validateType(typ); err != nil {
			return err
		}
	}
	finalizerLog.Info("started")
	for {
		select {
		case <-ticker.C:
			for _, typ := range f.types {
				if err := f.finalize(typ); err != nil {
					finalizerLog.Error(err, "unable to finalize subscription", "type", typ)
				}
			}
		case <-stop:
			finalizerLog.Info("stopped")
			return nil
		}
	}
}

func (f *subscriptionFinalizer) finalize(typ core_model.ResourceType) error {
	ctx := context.Background()
	insights, _ := registry.Global().NewList(typ)
	if err := f.rm.List(ctx, insights); err != nil {
		return err
	}

	for _, item := range insights.GetItems() {
		log := finalizerLog.WithValues("type", typ, "name", item.GetMeta().GetName(), "mesh", item.GetMeta().GetMesh())
		insight := item.GetSpec().(kuma_interfaces.Insight)
		if !insight.IsOnline() {
			continue
		}
		if insight.GetLastSubscription().GetCandidateForDisconnect() {
			log.V(1).Info("mark subscription as disconnected")
			insight.GetLastSubscription().SetDisconnectTime(core.Now())
		} else {
			log.V(1).Info("mark subscription as a candidate for disconnect")
			insight.GetLastSubscription().SetCandidateForDisconnect(true)
		}
		upsertInsight, _ := registry.Global().NewObject(typ)
		err := manager.Upsert(f.rm, core_model.MetaToResourceKey(item.GetMeta()), upsertInsight, func(_ core_model.Resource) bool {
			upsertInsight.GetSpec().(kuma_interfaces.Insight).UpdateSubscription(insight.GetLastSubscription())
			return true
		})
		if err != nil {
			log.Error(err, "unable to finalize subscription")
		}
	}
	return nil
}

func validateType(typ core_model.ResourceType) error {
	obj, err := registry.Global().NewObject(typ)
	if err != nil {
		return err
	}
	_, ok := obj.GetSpec().(kuma_interfaces.Insight)
	if !ok {
		return errors.Errorf("type %v doesn't implement interfaces.Insight", typ)
	}
	return nil
}

func (f *subscriptionFinalizer) NeedLeaderElection() bool {
	return true
}
