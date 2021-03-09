package context

import (
	"context"

	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/kds/mux"
	"github.com/kumahq/kuma/pkg/kds/reconcile"
	"github.com/kumahq/kuma/pkg/kds/util"
)

var log = core.Log.WithName("kds")

type Context struct {
	RemoteClientCtx       context.Context
	GlobalServerCallbacks []mux.Callbacks
	GlobalProvidedFilter  reconcile.ResourceFilter
	RemoteProvidedFilter  reconcile.ResourceFilter
}

func DefaultContext(manager manager.ResourceManager, zone string) *Context {
	return &Context{
		RemoteClientCtx:      context.Background(),
		GlobalProvidedFilter: GlobalProvidedFilter(manager),
		RemoteProvidedFilter: RemoteProvidedFilter(zone),
	}
}

// GlobalProvidedFilter returns ResourceFilter which filters Resources provided by Global, specifically
// excludes Dataplanes and Ingresses from 'clusterID' cluster
func GlobalProvidedFilter(rm manager.ResourceManager) reconcile.ResourceFilter {
	return func(clusterID string, r model.Resource) bool {
		if r.GetType() == system.ConfigType && r.GetMeta().GetName() != config_manager.ClusterIdConfigKey {
			return false
		}
		if r.GetType() != mesh.DataplaneType {
			return true
		}
		if !r.(*mesh.DataplaneResource).Spec.IsIngress() {
			return false
		}
		if clusterID == util.ZoneTag(r) {
			// don't need to sync resource to the zone where resource is originated from
			return false
		}
		zone := system.NewZoneResource()
		if err := rm.Get(context.Background(), zone, store.GetByKey(util.ZoneTag(r), model.NoMesh)); err != nil {
			log.Error(err, "failed to get zone", "zone", util.ZoneTag(r))
			// since there is no explicit 'enabled: false' then we don't
			// make any strong decisions which might affect connectivity
			return true
		}
		return zone.Spec.IsEnabled()
	}
}

// RemoteProvidedFilter filter Resources provided by Remote, specifically Ingresses that belongs to another zones
func RemoteProvidedFilter(clusterName string) reconcile.ResourceFilter {
	return func(_ string, r model.Resource) bool {
		if r.GetType() == mesh.DataplaneType {
			return clusterName == util.ZoneTag(r)
		}
		return r.GetType() == mesh.DataplaneInsightType
	}
}
