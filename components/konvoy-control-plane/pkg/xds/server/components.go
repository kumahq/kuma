package server

import (
	"context"
	"time"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"

	//core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
	util_watchdog "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/watchdog"
	xds_sync "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/sync"
	xds_template "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/template"

	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server"
)

func DefaultReconciler(rt core_runtime.Runtime) SnapshotReconciler {
	return &reconciler{
		&templateSnapshotGenerator{
			ProxyTemplateResolver: &simpleProxyTemplateResolver{
				ResourceManager:      rt.ResourceManager(),
				DefaultProxyTemplate: xds_template.DefaultProxyTemplate,
			},
		},
		&simpleSnapshotCacher{rt.XDS().Hasher(), rt.XDS().Cache()},
	}
}

func MatchDataplaneTrafficPermissions(dataplane *mesh_core.DataplaneResource, trafficPermissions *mesh_core.TrafficPermissionResourceList) *mesh_core.TrafficPermissionResourceList {
	matchedPerms := []*mesh_core.TrafficPermissionResource{}

	/*
		- For each Traffic Permission:
			- If there is a Destinations rule:
				- If the rule also has a Sources component:
					- include the Traffic Permission if the dataplane contains all of the
					  traffic permissions tags

		- TODO(gszr) ideally, the traffic permission should be statically validated to contain valid
		  rules (i.e., rules containing both Sources and Destinations). An error must be returned
		  in `kumactl apply` otherwise.
	*/

	for _, perm := range trafficPermissions.Items {
		matchedRules := []*mesh_proto.TrafficPermission_Rule{}

		for _, rule := range perm.Spec.Rules {
			for _, dest := range rule.Destinations {
				if len(rule.Sources) > 0 && dataplane.Spec.MatchTags(dest.Match) {
					matchedRules = append(matchedRules, &mesh_proto.TrafficPermission_Rule{
						Sources:      rule.Sources,
						Destinations: rule.Destinations,
					})
				}
			}
		}

		matchedPerms = append(matchedPerms, &mesh_core.TrafficPermissionResource{
			Meta: perm.Meta,
			Spec: mesh_proto.TrafficPermission{
				Rules: matchedRules,
			},
		})
	}

	return &mesh_core.TrafficPermissionResourceList{
		Items: matchedPerms,
	}
}

func DefaultDataplaneSyncTracker(rt core_runtime.Runtime, reconciler SnapshotReconciler) envoy_xds.Callbacks {
	return xds_sync.NewDataplaneSyncTracker(func(key core_model.ResourceKey) util_watchdog.Watchdog {
		log := xdsServerLog.WithName("dataplane-sync-watchdog").WithValues("dataplaneKey", key)
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(rt.Config().XdsServer.DataplaneConfigurationRefreshInterval)
			},
			OnTick: func() error {
				ctx := context.Background()
				dataplane := &mesh_core.DataplaneResource{}
				proxyId := xds.ProxyId{
					Name:      key.Name,
					Namespace: key.Namespace,
					Mesh:      key.Mesh,
				}

				if err := rt.ResourceManager().Get(ctx, dataplane, core_store.GetBy(key)); err != nil {
					if core_store.IsResourceNotFound(err) {
						return reconciler.Clear(&proxyId)
					}
					return err
				}

				// obtain all traffic permissions for the current mesh
				trafficPermissionsList := &mesh_core.TrafficPermissionResourceList{}
				if err := rt.ResourceManager().List(ctx, trafficPermissionsList, store.ListByMesh(key.Mesh)); err != nil {
					if core_store.IsResourceNotFound(err) {
						return reconciler.Clear(&proxyId)
					}
					return err
				}

				// match traffic permissions against the current dataplane object
				matchedPermissions := MatchDataplaneTrafficPermissions(dataplane, trafficPermissionsList)

				proxy := xds.Proxy{
					Id:                 proxyId,
					Dataplane:          dataplane,
					TrafficPermissions: matchedPermissions,
				}

				return reconciler.Reconcile(&proxy)
			},
			OnError: func(err error) {
				log.Error(err, "OnTick() failed")
			},
		}
	})
}

func DefaultDataplaneStatusTracker(rt core_runtime.Runtime) DataplaneStatusTracker {
	return NewDataplaneStatusTracker(rt, func(accessor SubscriptionStatusAccessor) DataplaneInsightSink {
		return NewDataplaneInsightSink(
			accessor,
			func() *time.Ticker {
				return time.NewTicker(rt.Config().XdsServer.DataplaneStatusFlushInterval)
			},
			NewDataplaneInsightStore(rt.ResourceManager()))
	})
}
