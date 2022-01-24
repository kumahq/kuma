package context

import (
	"context"
	"strings"

	"google.golang.org/protobuf/types/known/wrapperspb"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
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
	zone_tokens "github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
	"github.com/kumahq/kuma/pkg/util/rsa"
)

var log = core.Log.WithName("kds")

type Context struct {
	ZoneClientCtx        context.Context
	GlobalProvidedFilter reconcile.ResourceFilter
	ZoneProvidedFilter   reconcile.ResourceFilter
	GlobalServerFilters  []mux.Filter
	// Configs contains the names of system.ConfigResource that will be transferred from Global to Zone
	Configs map[string]bool

	GlobalResourceMapper reconcile.ResourceMapper
}

func DefaultContext(manager manager.ResourceManager, zone string) *Context {
	configs := map[string]bool{
		config_manager.ClusterIdConfigKey: true,
	}

	ctx := context.Background()

	return &Context{
		ZoneClientCtx:        ctx,
		GlobalProvidedFilter: GlobalProvidedFilter(manager, configs),
		ZoneProvidedFilter:   ZoneProvidedFilter(zone),
		Configs:              configs,
		GlobalResourceMapper: MapZoneTokenSigningKeyGlobalToPublicKey(ctx, manager),
	}
}

func MapZoneTokenSigningKeyGlobalToPublicKey(
	_ context.Context,
	_ manager.ResourceManager,
) reconcile.ResourceMapper {
	return func(r model.Resource) (model.Resource, error) {
		resType := r.Descriptor().Name
		currentMeta := r.GetMeta()
		resName := currentMeta.GetName()

		if resType == system.GlobalSecretType && strings.HasPrefix(resName, zone_tokens.SigningKeyPrefix) {
			signingKeyBytes := r.(*system.GlobalSecretResource).Spec.GetData().GetValue()
			publicKeyBytes, err := rsa.FromPrivateKeyPEMBytesToPublicKeyPEMBytes(signingKeyBytes)
			if err != nil {
				return nil, err
			}

			publicSigningKeyResource := system.NewGlobalSecretResource()
			newResName := strings.ReplaceAll(
				resName,
				zone_tokens.SigningKeyPrefix,
				zone_tokens.SigningPublicKeyPrefix,
			)
			publicSigningKeyResource.SetMeta(util.CloneResourceMetaWithNewName(currentMeta, newResName))

			if err := publicSigningKeyResource.SetSpec(&system_proto.Secret{
				Data: &wrapperspb.BytesValue{Value: publicKeyBytes},
			}); err != nil {
				return nil, err
			}

			return publicSigningKeyResource, nil
		}

		return r, nil
	}
}

// GlobalProvidedFilter returns ResourceFilter which filters Resources provided by Global, specifically
// excludes Dataplanes and Ingresses from 'clusterID' cluster
func GlobalProvidedFilter(rm manager.ResourceManager, configs map[string]bool) reconcile.ResourceFilter {
	return func(clusterID string, r model.Resource) bool {
		resType := r.Descriptor().Name
		resName := r.GetMeta().GetName()

		if resType == system.ConfigType && !configs[resName] {
			return false
		}
		if resType == system.GlobalSecretType {
			return util.ResourceNameHasAtLeastOneOfPrefixes(
				resName,
				zoneingress.ZoneIngressSigningKeyPrefix,
				zone_tokens.SigningKeyPrefix,
			)
		}
		if resType != mesh.DataplaneType && resType != mesh.ZoneIngressType {
			return true
		}
		if resType == mesh.DataplaneType {
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
