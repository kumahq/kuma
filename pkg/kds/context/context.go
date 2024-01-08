package context

import (
	"context"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	config_store "github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/hash"
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
	ZoneClientCtx         context.Context
	GlobalProvidedFilter  reconcile.ResourceFilter
	ZoneProvidedFilter    reconcile.ResourceFilter
	GlobalServerFilters   []mux.Filter
	GlobalServerFiltersV2 []mux.FilterV2
	// Configs contains the names of system.ConfigResource that will be transferred from Global to Zone
	Configs map[string]bool

	GlobalResourceMapper reconcile.ResourceMapper
	ZoneResourceMapper   reconcile.ResourceMapper

	EnvoyAdminRPCs           service.EnvoyAdminRPCs
	ServerStreamInterceptors []grpc.StreamServerInterceptor
	ServerUnaryInterceptor   []grpc.UnaryServerInterceptor
}

func DefaultContext(
	ctx context.Context,
	manager manager.ResourceManager,
	cfg kuma_cp.Config,
) *Context {
	configs := map[string]bool{
		config_manager.ClusterIdConfigKey: true,
	}

	globalMappers := []reconcile.ResourceMapper{
		UpdateResourceMeta(util.WithLabel(mesh_proto.ResourceOriginLabel, mesh_proto.ResourceOriginGlobal)),
		If(
			And(
				TypeIs(system.GlobalSecretType),
				NameHasPrefix(zone_tokens.SigningKeyPrefix),
			),
			MapZoneTokenSigningKeyGlobalToPublicKey),
		If(
			And(
				IsKubernetes(cfg.Store.Type),
				Not(TypeIs(system.ConfigType)), Not(TypeIs(system.SecretType)), Not(TypeIs(system.GlobalSecretType)),
			),
			RemoveK8sSystemNamespaceSuffixMapper(cfg.Store.Kubernetes.SystemNamespace)),
		If(
			And(
				ScopeIs(core_model.ScopeMesh),
				Not(TypeIs(system.SecretType)),
			),
			AddHashSuffix),
	}

	zoneMappers := []reconcile.ResourceMapper{
		MapInsightResourcesZeroGeneration,
		If(
			IsKubernetes(cfg.Store.Type),
			RemoveK8sSystemNamespaceSuffixMapper(cfg.Store.Kubernetes.SystemNamespace)),
	}

	return &Context{
		ZoneClientCtx:        ctx,
		GlobalProvidedFilter: GlobalProvidedFilter(manager, configs),
		ZoneProvidedFilter:   ZoneProvidedFilter,
		Configs:              configs,
		GlobalResourceMapper: CompositeResourceMapper(globalMappers...),
		ZoneResourceMapper:   CompositeResourceMapper(zoneMappers...),
		EnvoyAdminRPCs:       service.NewEnvoyAdminRPCs(),
	}
}

// CompositeResourceMapper combines the given ResourceMappers into
// a single ResourceMapper which calls each in order. If an error
// occurs, the first one is returned and no further mappers are executed.
func CompositeResourceMapper(mappers ...reconcile.ResourceMapper) reconcile.ResourceMapper {
	return func(features kds.Features, r core_model.Resource) (core_model.Resource, error) {
		var err error
		for _, mapper := range mappers {
			if mapper == nil {
				continue
			}

			r, err = mapper(features, r)
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
func MapInsightResourcesZeroGeneration(_ kds.Features, r core_model.Resource) (core_model.Resource, error) {
	if spec, ok := r.GetSpec().(specWithDiscoverySubscriptions); ok {
		spec = proto.Clone(spec).(specWithDiscoverySubscriptions)
		for _, sub := range spec.GetSubscriptions() {
			sub.Generation = 0
		}

		meta := r.GetMeta()
		resType := reflect.TypeOf(r).Elem()

		newR := reflect.New(resType).Interface().(core_model.Resource)
		newR.SetMeta(meta)
		if err := newR.SetSpec(spec.(core_model.ResourceSpec)); err != nil {
			panic(any(errors.Wrap(err, "error setting spec on resource")))
		}

		return newR, nil
	}

	return r, nil
}

func MapZoneTokenSigningKeyGlobalToPublicKey(_ kds.Features, r core_model.Resource) (core_model.Resource, error) {
	signingKeyBytes := r.(*system.GlobalSecretResource).Spec.GetData().GetValue()
	publicKeyBytes, err := rsa.FromPrivateKeyPEMBytesToPublicKeyPEMBytes(signingKeyBytes)
	if err != nil {
		return nil, err
	}

	publicSigningKeyResource := system.NewGlobalSecretResource()
	newResName := strings.ReplaceAll(
		r.GetMeta().GetName(),
		zone_tokens.SigningKeyPrefix,
		zone_tokens.SigningPublicKeyPrefix,
	)
	publicSigningKeyResource.SetMeta(util.CloneResourceMeta(r.GetMeta(), util.WithName(newResName)))

	if err := publicSigningKeyResource.SetSpec(&system_proto.Secret{
		Data: &wrapperspb.BytesValue{Value: publicKeyBytes},
	}); err != nil {
		return nil, err
	}

	return publicSigningKeyResource, nil
}

// RemoveK8sSystemNamespaceSuffixMapper is a mapper responsible for removing control plane system namespace suffixes
// from names of resources if resources are stored in kubernetes.
func RemoveK8sSystemNamespaceSuffixMapper(k8sSystemNamespace string) reconcile.ResourceMapper {
	return func(_ kds.Features, r core_model.Resource) (core_model.Resource, error) {
		util.TrimSuffixFromName(r, k8sSystemNamespace)
		return r, nil
	}
}

func TypeIs(rtype core_model.ResourceType) func(core_model.Resource) bool {
	return func(r core_model.Resource) bool {
		return r.Descriptor().Name == rtype
	}
}

func IsKubernetes(storeType config_store.StoreType) func(core_model.Resource) bool {
	return func(_ core_model.Resource) bool {
		return storeType == config_store.KubernetesStore
	}
}

func ScopeIs(s core_model.ResourceScope) func(core_model.Resource) bool {
	return func(r core_model.Resource) bool {
		return r.Descriptor().Scope == s
	}
}

func NameHasPrefix(prefix string) func(core_model.Resource) bool {
	return func(r core_model.Resource) bool {
		return strings.HasPrefix(r.GetMeta().GetName(), prefix)
	}
}

func Not(f func(core_model.Resource) bool) func(core_model.Resource) bool {
	return func(r core_model.Resource) bool {
		return !f(r)
	}
}

func And(fs ...func(core_model.Resource) bool) func(core_model.Resource) bool {
	return func(r core_model.Resource) bool {
		for _, f := range fs {
			if !f(r) {
				return false
			}
		}
		return true
	}
}

func If(condition func(core_model.Resource) bool, m reconcile.ResourceMapper) reconcile.ResourceMapper {
	return func(features kds.Features, r core_model.Resource) (core_model.Resource, error) {
		if condition(r) {
			return m(features, r)
		}
		return r, nil
	}
}

// AddHashSuffix is a mapper responsible for adding hash suffix to the name of the resource
func AddHashSuffix(features kds.Features, r core_model.Resource) (core_model.Resource, error) {
	if !features.HasFeature(kds.FeatureHashSuffix) {
		return r, nil
	}

	newObj := r.Descriptor().NewObject()
	newMeta := util.CloneResourceMeta(r.GetMeta(), util.WithName(hash.HashedName(r.GetMeta().GetMesh(), core_model.GetDisplayName(r))))
	newObj.SetMeta(newMeta)
	_ = newObj.SetSpec(r.GetSpec())

	return newObj, nil
}

func UpdateResourceMeta(fs ...util.CloneResourceMetaOpt) reconcile.ResourceMapper {
	return func(_ kds.Features, r core_model.Resource) (core_model.Resource, error) {
		r.SetMeta(util.CloneResourceMeta(r.GetMeta(), fs...))
		return r, nil
	}
}

func GlobalProvidedFilter(rm manager.ResourceManager, configs map[string]bool) reconcile.ResourceFilter {
	return func(ctx context.Context, clusterID string, features kds.Features, r core_model.Resource) bool {
		resName := r.GetMeta().GetName()

		switch {
		case r.Descriptor().Name == system.ConfigType:
			return configs[resName]
		case r.Descriptor().Name == system.GlobalSecretType:
			return util.ResourceNameHasAtLeastOneOfPrefixes(resName, []string{
				zoneingress.ZoneIngressSigningKeyPrefix,
				zone_tokens.SigningKeyPrefix,
			}...)
		case r.Descriptor().KDSFlags.Has(core_model.GlobalToAllButOriginalZoneFlag):
			zoneTag := util.ZoneTag(r)

			if clusterID == zoneTag {
				// don't need to sync resource to the zone where resource is originated from
				return false
			}

			zone := system.NewZoneResource()
			if err := rm.Get(ctx, zone, store.GetByKey(zoneTag, core_model.NoMesh)); err != nil {
				log.Error(err, "failed to get zone", "zone", zoneTag)
				// since there is no explicit 'enabled: false' then we don't
				// make any strong decisions which might affect connectivity
				return true
			}

			return zone.Spec.IsEnabled()
		default:
			return core_model.IsLocallyOriginated(config_core.Global, r)
		}
	}
}

func ZoneProvidedFilter(_ context.Context, _ string, _ kds.Features, r core_model.Resource) bool {
	return core_model.IsLocallyOriginated(config_core.Zone, r)
}
