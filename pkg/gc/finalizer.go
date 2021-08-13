package gc

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/api/generic"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

var (
	finalizerLog = core.Log.WithName("finalizer")
)

// Every Insight has a statusSink that periodically changes ResourceVersion while Kuma CP <-> DPP
// connection is active. This updates happens every <KumaCP.Config>.Metrics.Dataplane.IdleTimeout / 2
// even if no real updates of the resource happen.
//
// subscriptionFinalizer is a component that allows to finalize subscriptions for Insights. The component iterates
// over all online Insights and checks that ResourceVersion was changed since the last check. This check
// happens every <KumaCP.Config>.Metrics.Dataplane.IdleTimeout. If ResourceVersion didn't change then component
// finalizes the subscription by setting DisconnectTime.
//
// This component allows to solve the corner case we had with Insights:
// 1. Kuma CP is down
// 2. DPP is down
// 3. Kuma CP is up
// 4. DPP status is Online whereas it should be Offline

type insightMap map[core_model.ResourceKey]string

type insightsByType map[core_model.ResourceType]insightMap

type subscriptionFinalizer struct {
	rm        manager.ResourceManager
	newTicker func() *time.Ticker
	types     []core_model.ResourceType
	insights  insightsByType
}

func NewSubscriptionFinalizer(rm manager.ResourceManager, newTicker func() *time.Ticker, types ...core_model.ResourceType) (component.Component, error) {
	insights := insightsByType{}
	for _, typ := range types {
		if !isInsightType(typ) {
			return nil, errors.Errorf("%q type is not an Insight", typ)
		}
		insights[typ] = map[core_model.ResourceKey]string{}
	}

	return &subscriptionFinalizer{
		rm:        rm,
		types:     types,
		newTicker: newTicker,
		insights:  insights,
	}, nil
}

func (f *subscriptionFinalizer) Start(stop <-chan struct{}) error {
	ticker := f.newTicker()
	defer ticker.Stop()

	finalizerLog.Info("started")
	for {
		select {
		case <-ticker.C:
			for _, typ := range f.types {
				if err := f.checkResourceVersion(typ); err != nil {
					finalizerLog.Error(err, "unable to check insight's resourceVersion", "type", typ)
				}
			}
		case <-stop:
			finalizerLog.Info("stopped")
			return nil
		}
	}
}

func (f *subscriptionFinalizer) checkResourceVersion(typ core_model.ResourceType) error {
	// get all the insights for provided type
	insights, _ := registry.Global().NewList(typ)
	if err := f.rm.List(context.Background(), insights); err != nil {
		return err
	}

	// delete items from the map that don't exist in the upstream
	f.removeDeletedInsights(insights)

	for _, item := range insights.GetItems() {
		log := finalizerLog.WithValues("type", typ, "name", item.GetMeta().GetName(), "mesh", item.GetMeta().GetMesh())
		key := core_model.MetaToResourceKey(item.GetMeta())
		insight := item.GetSpec().(generic.Insight)

		if !insight.IsOnline() {
			delete(f.insights[typ], key)
			continue
		}

		oldVersion, ok := f.insights[typ][key]
		if !ok || oldVersion != item.GetMeta().GetVersion() {
			// resourceVersion changed since the last check, don't finalize
			// the subscription, update map with fresh data
			f.insights[typ][key] = item.GetMeta().GetVersion()
			continue
		}

		log.V(1).Info("mark subscription as disconnected")
		insight.GetLastSubscription().SetDisconnectTime(core.Now())

		upsertInsight, _ := registry.Global().NewObject(typ)
		err := manager.Upsert(f.rm, key, upsertInsight, func(r core_model.Resource) error {
			return upsertInsight.GetSpec().(generic.Insight).UpdateSubscription(insight.GetLastSubscription())
		})
		if err != nil {
			log.Error(err, "unable to finalize subscription")
			return err
		}
		delete(f.insights[typ], key)
	}
	return nil
}

func (f *subscriptionFinalizer) removeDeletedInsights(insights core_model.ResourceList) {
	byResourceKey := map[core_model.ResourceKey]bool{}
	for _, item := range insights.GetItems() {
		byResourceKey[core_model.MetaToResourceKey(item.GetMeta())] = true
	}
	for rk := range f.insights[insights.GetItemType()] {
		if !byResourceKey[rk] {
			delete(f.insights[insights.GetItemType()], rk)
		}
	}
}

func isInsightType(typ core_model.ResourceType) bool {
	obj, err := registry.Global().NewObject(typ)
	if err != nil {
		panic(err)
	}
	_, ok := obj.GetSpec().(generic.Insight)
	return ok
}

func (f *subscriptionFinalizer) NeedLeaderElection() bool {
	return true
}
