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
	"github.com/kumahq/kuma/pkg/kds/reconcile"
	"github.com/kumahq/kuma/pkg/kds/util"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
)

var log = core.Log.WithName("kds")

type Context struct {
	ZoneClientCtx        context.Context
	GlobalProvidedFilter reconcile.ResourceFilter
	ZoneProvidedFilter   reconcile.ResourceFilter
	// Configs contains the names of system.ConfigResource that will be transferred from Global to Zone
	Configs map[string]bool
}

func DefaultContext(manager manager.ResourceManager, zone string) *Context {
	configs := map[string]bool{
		config_manager.ClusterIdConfigKey: true,
	}
	return &Context{
		ZoneClientCtx:        context.Background(),
		GlobalProvidedFilter: GlobalProvidedFilter(manager, configs),
		ZoneProvidedFilter:   ZoneProvidedFilter(zone),
		Configs:              configs,
	}
}

// GlobalProvidedFilter returns ResourceFilter which filters Resources provided by Global, specifically
// excludes Dataplanes and Ingresses from 'clusterID' cluster
func GlobalProvidedFilter(rm manager.ResourceManager, configs map[string]bool) reconcile.ResourceFilter {
	return func(clusterID string, r model.Resource) bool {
		resType := r.Descriptor().Name
		if resType == mesh.ZoneIngressType {
			return r.(*mesh.ZoneIngressResource).Spec.GetZone() != clusterID
		}
		if resType == system.ConfigType && !configs[r.GetMeta().GetName()] {
			return false
		}
		if resType == system.GlobalSecretType {
			return zoneingress.IsSigningKeyResource(model.MetaToResourceKey(r.GetMeta()))
		}
		if resType != mesh.DataplaneType {
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

// ZoneProvidedFilter filter Resources provided by Zone, specifically Ingresses that belongs to another zones
func ZoneProvidedFilter(clusterName string) reconcile.ResourceFilter {
	return func(_ string, r model.Resource) bool {
		resType := r.Descriptor().Name
		if resType == mesh.DataplaneType {
			return clusterName == util.ZoneTag(r)
		}
		if resType == mesh.DataplaneInsightType {
			return true
		}
		if resType == mesh.ZoneIngressType && !r.(*mesh.ZoneIngressResource).IsRemoteIngress(clusterName) {
			return true
		}
		if resType == mesh.ZoneIngressInsightType {
			return true
		}
		return false
	}
}
