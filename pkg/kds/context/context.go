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
	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/hash"
	"github.com/kumahq/kuma/pkg/kds/mux"
	"github.com/kumahq/kuma/pkg/kds/service"
	"github.com/kumahq/kuma/pkg/kds/util"
	reconcile_v2 "github.com/kumahq/kuma/pkg/kds/v2/reconcile"
	zone_tokens "github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/pkg/util/rsa"
	"github.com/kumahq/kuma/pkg/version"
)

const VersionHeader = "version"

var log = core.Log.WithName("kds")

type Context struct {
	ZoneClientCtx         context.Context
	GlobalProvidedFilter  reconcile_v2.ResourceFilter
	ZoneProvidedFilter    reconcile_v2.ResourceFilter
	GlobalServerFiltersV2 []mux.FilterV2
	// Configs contains the names of system.ConfigResource that will be transferred from Global to Zone
	Configs map[string]bool

	GlobalResourceMapper reconcile_v2.ResourceMapper
	ZoneResourceMapper   reconcile_v2.ResourceMapper

	EnvoyAdminRPCs           service.EnvoyAdminRPCs
	ServerStreamInterceptors []grpc.StreamServerInterceptor
	ServerUnaryInterceptor   []grpc.UnaryServerInterceptor
	CreateZoneOnFirstConnect bool
}

func DefaultContext(
	ctx context.Context,
	manager manager.ResourceManager,
	cfg kuma_cp.Config,
) *Context {
	configs := map[string]bool{
		config_manager.ClusterIdConfigKey: true,
	}

	globalMappers := []reconcile_v2.ResourceMapper{
		UpdateResourceMeta(util.WithLabel(mesh_proto.ResourceOriginLabel, string(mesh_proto.GlobalResourceOrigin))),
		reconcile_v2.If(
			reconcile_v2.And(
				reconcile_v2.TypeIs(system.GlobalSecretType),
				reconcile_v2.NameHasPrefix(zone_tokens.SigningKeyPrefix),
			),
			MapZoneTokenSigningKeyGlobalToPublicKey),
		reconcile_v2.If(
			reconcile_v2.IsKubernetes(cfg.Store.Type),
			RemoveK8sSystemNamespaceSuffixMapper(cfg.Store.Kubernetes.SystemNamespace)),
		reconcile_v2.If(
			reconcile_v2.TypeIs(meshservice_api.MeshServiceType),
			RemoveStatus()),
		reconcile_v2.If(
			reconcile_v2.And(
				// secrets already named with mesh prefix for uniqueness on k8s, also Zone CP expects secret names to be in
				// particular format to be able to reference them
				reconcile_v2.Not(reconcile_v2.TypeIs(system.SecretType)),
				// Zone CP expects secret names to be in particular format to be able to reference them
				reconcile_v2.Not(reconcile_v2.TypeIs(system.GlobalSecretType)),
				// Zone CP expects secret names to be in particular format to be able to reference them
				reconcile_v2.Not(reconcile_v2.TypeIs(system.ConfigType)),
				// Mesh name has to be the same. In multizone deployments it can only be applied on Global CP,
				// so we won't hit conflicts.
				reconcile_v2.Not(reconcile_v2.TypeIs(core_mesh.MeshType)),
			),
			HashSuffixMapper(true, mesh_proto.ZoneTag, mesh_proto.KubeNamespaceTag)),
	}

	zoneMappers := []reconcile_v2.ResourceMapper{
		UpdateResourceMeta(
			util.WithLabel(mesh_proto.ResourceOriginLabel, string(mesh_proto.ZoneResourceOrigin)),
			util.WithLabel(mesh_proto.ZoneTag, cfg.Multizone.Zone.Name),
			util.WithoutLabel(mesh_proto.DeletionGracePeriodStartedLabel),
		),
		MapInsightResourcesZeroGeneration,
		reconcile_v2.If(
			reconcile_v2.IsKubernetes(cfg.Store.Type),
			RemoveK8sSystemNamespaceSuffixMapper(cfg.Store.Kubernetes.SystemNamespace)),
		HashSuffixMapper(false, mesh_proto.ZoneTag, mesh_proto.KubeNamespaceTag),
	}
	ctx = metadata.AppendToOutgoingContext(ctx, VersionHeader, version.Build.Version)

	return &Context{
		ZoneClientCtx: ctx,
		GlobalProvidedFilter: CompositeResourceFilters(
			GlobalProvidedFilter(manager, configs),
			SkipUnsupportedHostnameGenerator,
		),
		ZoneProvidedFilter:       ZoneProvidedFilter,
		Configs:                  configs,
		GlobalResourceMapper:     CompositeResourceMapper(globalMappers...),
		ZoneResourceMapper:       CompositeResourceMapper(zoneMappers...),
		EnvoyAdminRPCs:           service.NewEnvoyAdminRPCs(),
		CreateZoneOnFirstConnect: true,
	}
}

func CompositeResourceFilters(filters ...reconcile_v2.ResourceFilter) reconcile_v2.ResourceFilter {
	return func(ctx context.Context, clusterID string, features kds.Features, r core_model.Resource) bool {
		for _, filter := range filters {
			if !filter(ctx, clusterID, features, r) {
				return false
			}
		}
		return true
	}
}

// CompositeResourceMapper combines the given ResourceMappers into
// a single ResourceMapper which calls each in order. If an error
// occurs, the first one is returned and no further mappers are executed.
func CompositeResourceMapper(mappers ...reconcile_v2.ResourceMapper) reconcile_v2.ResourceMapper {
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
func RemoveK8sSystemNamespaceSuffixMapper(k8sSystemNamespace string) reconcile_v2.ResourceMapper {
	return func(_ kds.Features, r core_model.Resource) (core_model.Resource, error) {
		dotSuffix := fmt.Sprintf(".%s", k8sSystemNamespace)
		newName := strings.TrimSuffix(r.GetMeta().GetName(), dotSuffix)
		return util.CloneResource(r, util.WithResourceName(newName)), nil
	}
}

func RemoveStatus() reconcile_v2.ResourceMapper {
	return func(_ kds.Features, r core_model.Resource) (core_model.Resource, error) {
		return util.CloneResource(r, util.WithoutStatus()), nil
	}
}

// HashSuffixMapper returns mapper that adds a hash suffix to the name during KDS sync
func HashSuffixMapper(checkKDSFeature bool, labelsToUse ...string) reconcile_v2.ResourceMapper {
	return func(features kds.Features, r core_model.Resource) (core_model.Resource, error) {
		if checkKDSFeature && !features.HasFeature(kds.FeatureHashSuffix) {
			return r, nil
		}

		name := core_model.GetDisplayName(r.GetMeta())
		values := make([]string, 0, len(labelsToUse))
		for _, lbl := range labelsToUse {
			labelValue, ok := r.GetMeta().GetLabels()[lbl]
			if ok {
				values = append(values, labelValue)
			}
		}

		return util.CloneResource(r, util.WithResourceName(hash.HashedName(r.GetMeta().GetMesh(), name, values...))), nil
	}
}

func UpdateResourceMeta(fs ...util.CloneResourceMetaOpt) reconcile_v2.ResourceMapper {
	return func(_ kds.Features, r core_model.Resource) (core_model.Resource, error) {
		newRes := util.CloneResource(r)
		newRes.SetMeta(util.CloneResourceMeta(r.GetMeta(), fs...))
		return newRes, nil
	}
}

func GlobalProvidedFilter(rm manager.ResourceManager, configs map[string]bool) reconcile_v2.ResourceFilter {
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
			// TODO (Icarus9913): replace the function by model.ZoneOfResource(r)
			// Reference: https://github.com/kumahq/kuma/issues/10952
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
			return core_model.IsLocallyOriginated(config_core.Global, r.GetMeta().GetLabels())
		}
	}
}

func SkipUnsupportedHostnameGenerator(ctx context.Context, clusterID string, features kds.Features, r core_model.Resource) bool {
	if r.Descriptor().Name != hostnamegenerator_api.HostnameGeneratorType {
		return true
	}
	if !features.HasFeature(kds.FeatureHostnameGeneratorMzSelector) {
		if r.GetSpec().(*hostnamegenerator_api.HostnameGenerator).Selector.MeshMultiZoneService != nil {
			return false
		}
	}
	return true
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
	return core_model.IsLocallyOriginated(config_core.Zone, r.GetMeta().GetLabels()) || r.Descriptor().KDSFlags == core_model.ZoneToGlobalFlag
}
