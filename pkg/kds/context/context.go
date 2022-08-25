package context

import (
	"context"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/mux"
	"github.com/kumahq/kuma/pkg/kds/reconcile"
	"github.com/kumahq/kuma/pkg/kds/service"
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
	ZoneResourceMapper   reconcile.ResourceMapper

	EnvoyAdminRPCs service.EnvoyAdminRPCs
}

func DefaultContext(ctx context.Context, manager manager.ResourceManager, zone string) *Context {
	configs := map[string]bool{
		config_manager.ClusterIdConfigKey: true,
	}

	return &Context{
		ZoneClientCtx:        ctx,
		GlobalProvidedFilter: GlobalProvidedFilter(manager, configs),
		ZoneProvidedFilter:   ZoneProvidedFilter(zone),
		Configs:              configs,
		GlobalResourceMapper: MapZoneTokenSigningKeyGlobalToPublicKey(ctx, manager),
		ZoneResourceMapper:   MapInsightResourcesZeroGeneration,
		EnvoyAdminRPCs:       service.NewEnvoyAdminRPCs(),
	}
}

// CompositeResourceMapper combines the given ResourceMappers into
// a single ResourceMapper which calls each in order. If an error
// occurs, the first one is returned and no further mappers are executed.
func CompositeResourceMapper(mappers ...reconcile.ResourceMapper) reconcile.ResourceMapper {
	return func(r model.Resource) (model.Resource, error) {
		var err error
		for _, mapper := range mappers {
			r, err = mapper(r)
			if err != nil {
				return r, err
			}
		}
		return r, nil
	}
}

type specWithDiscoverySubscriptions interface {
	GetSubscriptions() []*mesh_proto.DiscoverySubscription
	ProtoReflect() protoreflect.Message
}

// MapInsightResourcesZeroGeneration zeros "generation" field in resources for which
// the field has only local relevance. This prevents reconciliation from unnecessarily
// deeming the object to have changed.
func MapInsightResourcesZeroGeneration(r model.Resource) (model.Resource, error) {
	if spec, ok := r.GetSpec().(specWithDiscoverySubscriptions); ok {
		spec = proto.Clone(spec).(specWithDiscoverySubscriptions)
		for _, sub := range spec.GetSubscriptions() {
			sub.Generation = 0
		}

		meta := r.GetMeta()
		resType := reflect.TypeOf(r).Elem()

		newR := reflect.New(resType).Interface().(model.Resource)
		newR.SetMeta(meta)
		if err := newR.SetSpec(spec.(model.ResourceSpec)); err != nil {
			panic(any(errors.Wrap(err, "error setting spec on resource")))
		}

		return newR, nil
	}

	return r, nil
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
// excludes Dataplanes, Ingresses and Egresses from 'clusterID' cluster
func GlobalProvidedFilter(rm manager.ResourceManager, configs map[string]bool) reconcile.ResourceFilter {
	return func(ctx context.Context, clusterID string, features kds.Features, r model.Resource) bool {
		resName := r.GetMeta().GetName()

		switch r.Descriptor().Name {
		case mesh.DataplaneType:
			return false
		case system.ConfigType:
			return configs[resName]
		case system.GlobalSecretType:
			prefixes := []string{zoneingress.ZoneIngressSigningKeyPrefix}
			if features.HasFeature(kds.FeatureZoneToken) {
				// We need to sync Zone Token signing keys only to zone cps that can support it.
				// Otherwise, Zone CP after the restart of either Zone or Global CP tries to recreate resource
				// The result is that it NACKs the DiscoveryResponse and gets into a loop.
				prefixes = append(prefixes, zone_tokens.SigningKeyPrefix)
			}
			return util.ResourceNameHasAtLeastOneOfPrefixes(resName, prefixes...)
		case mesh.ZoneIngressType:
			zoneTag := util.ZoneTag(r)

			if clusterID == zoneTag {
				// don't need to sync resource to the zone where resource is originated from
				return false
			}

			zone := system.NewZoneResource()
			if err := rm.Get(ctx, zone, store.GetByKey(zoneTag, model.NoMesh)); err != nil {
				log.Error(err, "failed to get zone", "zone", zoneTag)
				// since there is no explicit 'enabled: false' then we don't
				// make any strong decisions which might affect connectivity
				return true
			}

			return zone.Spec.IsEnabled()
		default:
			return true
		}
	}
}

// ZoneProvidedFilter filter Resources provided by Zone, specifically Ingresses
// that belongs to another zones
func ZoneProvidedFilter(clusterName string) reconcile.ResourceFilter {
	return func(_ context.Context, _ string, _ kds.Features, r model.Resource) bool {
		switch r.Descriptor().Name {
		case mesh.DataplaneType:
			return clusterName == util.ZoneTag(r)
		case mesh.ZoneIngressType:
			return !r.(*mesh.ZoneIngressResource).IsRemoteIngress(clusterName)
		case mesh.DataplaneInsightType,
			mesh.ZoneIngressInsightType,
			mesh.ZoneEgressType,
			mesh.ZoneEgressInsightType:
			return true
		default:
			return false
		}
	}
}
