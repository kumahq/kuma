package gc

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/api/generic"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/multitenant"
)

var finalizerLog = core.Log.WithName("finalizer")

// Every Insight has a statusSink that periodically increments the 'generation' counter while Kuma CP <-> DPP
// connection is active. This updates happens every <KumaCP.Config>.Metrics.Dataplane.IdleTimeout / 2.
//
// subscriptionFinalizer is a component that allows to finalize subscriptions for Insights. The component iterates
// over all online Insights and checks that 'generation' counter was changed since the last check. This check
// happens every <KumaCP.Config>.Metrics.Dataplane.IdleTimeout. If 'generation' counter didn't change then component
// finalizes the subscription by setting DisconnectTime.
//
// This component allows to solve the corner case we had with Insights:
// 1. Kuma CP is down
// 2. DPP is down
// 3. Kuma CP is up
// 4. DPP status is Online whereas it should be Offline

type descriptor struct {
	id         string
	generation uint32
}

type tenantID string

type insightMap map[core_model.ResourceKey]*descriptor

type insightsByType map[core_model.ResourceType]insightMap

type tenantInsights map[tenantID]insightsByType

type subscriptionFinalizer struct {
	rm        manager.ResourceManager
	newTicker func() *time.Ticker
	types     []core_model.ResourceType
	insights  tenantInsights
	tenants   multitenant.Tenants
	metric    prometheus.Summary
}

func NewSubscriptionFinalizer(
	rm manager.ResourceManager,
	tenants multitenant.Tenants,
	newTicker func() *time.Ticker,
	metrics core_metrics.Metrics,
	types ...core_model.ResourceType,
) (component.Component, error) {
	for _, typ := range types {
		if !isInsightType(typ) {
			return nil, errors.Errorf("%q type is not an Insight", typ)
		}
	}

	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_sub_finalizer",
		Help:       "Summary of Subscription finalizer component interval",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}

	return &subscriptionFinalizer{
		rm:        rm,
		types:     types,
		newTicker: newTicker,
		insights:  tenantInsights{},
		tenants:   tenants,
		metric:    metric,
	}, nil
}

func (f *subscriptionFinalizer) Start(stop <-chan struct{}) error {
	ticker := f.newTicker()
	defer ticker.Stop()

	finalizerLog.Info("started")
	for {
		select {
		case now := <-ticker.C:
			start := core.Now()
			tenantIds, err := f.tenants.GetIDs(context.TODO())
			if err != nil {
				finalizerLog.Error(err, "could not get contexts")
				break
			}
			for _, typ := range f.types {
				for _, tenantId := range tenantIds {
					if _, found := f.insights[tenantID(tenantId)]; !found {
						f.insights[tenantID(tenantId)] = insightsByType{}
					}
					if _, found := f.insights[tenantID(tenantId)][typ]; !found {
						f.insights[tenantID(tenantId)][typ] = insightMap{}
					}
					ctx := multitenant.WithTenant(context.TODO(), tenantId)
					if err := f.checkGeneration(user.Ctx(ctx, user.ControlPlane), typ, now); err != nil {
						finalizerLog.Error(err, "unable to check subscription's generation", "type", typ)
					}
				}
			}
			f.metric.Observe(float64(core.Now().Sub(start).Milliseconds()))
		case <-stop:
			finalizerLog.Info("stopped")
			return nil
		}
	}
}

func (f *subscriptionFinalizer) checkGeneration(ctx context.Context, typ core_model.ResourceType, now time.Time) error {
	// get all the insights for provided type
	insights, _ := registry.Global().NewList(typ)
	if err := f.rm.List(ctx, insights); err != nil {
		return err
	}

	tenantId, ok := multitenant.TenantFromCtx(ctx)
	if !ok {
		return multitenant.TenantMissingErr
	}

	// delete items from the map that don't exist in the upstream
	f.removeDeletedInsights(insights, tenantId)

	for _, item := range insights.GetItems() {
		log := finalizerLog.WithValues("type", typ, "name", item.GetMeta().GetName(), "mesh", item.GetMeta().GetMesh())
		key := core_model.MetaToResourceKey(item.GetMeta())
		insight := item.GetSpec().(generic.Insight)

		if !insight.IsOnline() {
			delete(f.insights[tenantID(tenantId)][typ], key)
			continue
		}

		old, ok := f.insights[tenantID(tenantId)][typ][key]
		if !ok || old.id != insight.GetLastSubscription().GetId() || old.generation != insight.GetLastSubscription().GetGeneration() {
			// something changed since the last check, either subscriptionId or generation were updated
			// don't finalize the subscription, update map with fresh data
			f.insights[tenantID(tenantId)][typ][key] = &descriptor{
				id:         insight.GetLastSubscription().GetId(),
				generation: insight.GetLastSubscription().GetGeneration(),
			}
			continue
		}

		log.V(1).Info("mark subscription as disconnected")
		insight.GetLastSubscription().SetDisconnectTime(now)

		upsertInsight, _ := registry.Global().NewObject(typ)
		err := manager.Upsert(ctx, f.rm, key, upsertInsight, func(r core_model.Resource) error {
			return upsertInsight.GetSpec().(generic.Insight).UpdateSubscription(insight.GetLastSubscription())
		})
		if err != nil {
			log.Error(err, "unable to finalize subscription")
			return err
		}
		delete(f.insights[tenantID(tenantId)][typ], key)
	}
	return nil
}

func (f *subscriptionFinalizer) removeDeletedInsights(insights core_model.ResourceList, tenantId string) {
	byResourceKey := map[core_model.ResourceKey]bool{}
	for _, item := range insights.GetItems() {
		byResourceKey[core_model.MetaToResourceKey(item.GetMeta())] = true
	}
	for rk := range f.insights[tenantID(tenantId)][insights.GetItemType()] {
		if !byResourceKey[rk] {
			delete(f.insights[tenantID(tenantId)][insights.GetItemType()], rk)
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
