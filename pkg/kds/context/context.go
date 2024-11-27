package context

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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
	"github.com/kumahq/kuma/pkg/util/rsa"
	"github.com/kumahq/kuma/pkg/version"
)

const VersionHeader = "version"

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
		UpdateResourceMeta(util.WithLabel(mesh_proto.ResourceOriginLabel, string(mesh_proto.GlobalResourceOrigin))),
		reconcile.If(
			reconcile.And(
				reconcile.TypeIs(system.GlobalSecretType),
				reconcile.NameHasPrefix(zone_tokens.SigningKeyPrefix),
			),
			MapZoneTokenSigningKeyGlobalToPublicKey),
		reconcile.If(
			reconcile.IsKubernetes(cfg.Store.Type),
			RemoveK8sSystemNamespaceSuffixMapper(cfg.Store.Kubernetes.SystemNamespace)),
		reconcile.If(
			reconcile.And(
				reconcile.ScopeIs(core_model.ScopeMesh),
				// secrets already named with mesh prefix for uniqueness on k8s, also Zone CP expects secret names to be in
				// particular format to be able to reference them
				reconcile.Not(reconcile.TypeIs(system.SecretType)),
			),
			HashSuffixMapper(true)),
	}

	zoneMappers := []reconcile.ResourceMapper{
		UpdateResourceMeta(
			util.WithLabel(mesh_proto.ResourceOriginLabel, string(mesh_proto.ZoneResourceOrigin)),
			util.WithLabel(mesh_proto.ZoneTag, cfg.Multizone.Zone.Name),
		),
		MapInsightResourcesZeroGeneration,
		reconcile.If(
			reconcile.IsKubernetes(cfg.Store.Type),
			RemoveK8sSystemNamespaceSuffixMapper(cfg.Store.Kubernetes.SystemNamespace)),
		HashSuffixMapper(false, mesh_proto.ZoneTag, mesh_proto.KubeNamespaceTag),
	}
	ctx = metadata.AppendToOutgoingContext(ctx, VersionHeader, version.Build.Version)

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
		newObj := r.Descriptor().NewObject()
		dotSuffix := fmt.Sprintf(".%s", k8sSystemNamespace)
		newName := strings.TrimSuffix(r.GetMeta().GetName(), dotSuffix)
		newMeta := util.CloneResourceMeta(r.GetMeta(), util.WithName(newName))
		newObj.SetMeta(newMeta)
		_ = newObj.SetSpec(r.GetSpec())
		return newObj, nil
	}
}

// HashSuffixMapper returns mapper that adds a hash suffix to the name during KDS sync
func HashSuffixMapper(checkKDSFeature bool, labelsToUse ...string) reconcile.ResourceMapper {
	return func(features kds.Features, r core_model.Resource) (core_model.Resource, error) {
		if checkKDSFeature && !features.HasFeature(kds.FeatureHashSuffix) {
			return r, nil
		}

		name := core_model.GetDisplayName(r.GetMeta())
		values := make([]string, 0, len(labelsToUse))
		for _, lbl := range labelsToUse {
			values = append(values, r.GetMeta().GetLabels()[lbl])
		}

		newObj := r.Descriptor().NewObject()
		newMeta := util.CloneResourceMeta(r.GetMeta(), util.WithName(hash.HashedName(r.GetMeta().GetMesh(), name, values...)))
		newObj.SetMeta(newMeta)
		_ = newObj.SetSpec(r.GetSpec())

		return newObj, nil
	}
}

func UpdateResourceMeta(fs ...util.CloneResourceMetaOpt) reconcile.ResourceMapper {
	return func(_ kds.Features, r core_model.Resource) (core_model.Resource, error) {
		newObj := r.Descriptor().NewObject()
		newMeta := util.CloneResourceMeta(r.GetMeta(), fs...)
		newObj.SetMeta(newMeta)
		_ = newObj.SetSpec(r.GetSpec())

		return newObj, nil
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

func ZoneProvidedFilter(_ context.Context, localZone string, _ kds.Features, r core_model.Resource) bool {
	if zi, ok := r.(*core_mesh.ZoneIngressResource); ok {
		// Old zones don't have a 'kuma.io/zone' label on ZoneIngress, when upgrading to the new 2.6 version
		// we don't want Zone CP to sync ZoneIngresses without 'kuma.io/zone' label to Global pretending
		// they're originating here. That's why upgrade from 2.5 to 2.6 (and 2.7) requires casting resource
		// to *core_mesh.ZoneIngressResource and checking its 'spec.zone' field.
		// todo: remove in 2 releases after 2.6.x
		return !zi.IsRemoteIngress(localZone)
	}
	return core_model.IsLocallyOriginated(config_core.Zone, r) || r.Descriptor().KDSFlags == core_model.ZoneToGlobalFlag
}
