package context

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/wrapperspb"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/kri"
	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/hash"
	"github.com/kumahq/kuma/pkg/kds/mux"
	"github.com/kumahq/kuma/pkg/kds/service"
	"github.com/kumahq/kuma/pkg/kds/util"
	reconcile_v2 "github.com/kumahq/kuma/pkg/kds/v2/reconcile"
	"github.com/kumahq/kuma/pkg/util/pointer"
	"github.com/kumahq/kuma/pkg/util/rsa"
	"github.com/kumahq/kuma/pkg/version"
)

const VersionHeader = "version"

var log = core.Log.WithName("kds")

type Context struct {
	ZoneClientCtx         context.Context
	TypesSentByZone       []core_model.ResourceType
	TypesSentByGlobal     []core_model.ResourceType
	GlobalProvidedFilter  reconcile_v2.ResourceFilter
	ZoneProvidedFilter    reconcile_v2.ResourceFilter
	GlobalServerFiltersV2 []mux.FilterV2

	GlobalResourceMapper reconcile_v2.ResourceMapper
	ZoneResourceMapper   reconcile_v2.ResourceMapper

	EnvoyAdminRPCs           service.EnvoyAdminRPCs
	ServerStreamInterceptors []grpc.StreamServerInterceptor
	ServerUnaryInterceptor   []grpc.UnaryServerInterceptor
	CreateZoneOnFirstConnect bool
}

// KDSSyncedConfigs A select set of keys that are synced from global to zones
var KDSSyncedConfigs = map[string]struct{}{
	config_manager.ClusterIdConfigKey: {},
}

func DefaultContext(
	ctx context.Context,
	manager manager.ResourceManager,
	cfg kuma_cp.Config,
) *Context {
	globalMappers := []reconcile_v2.ResourceMapper{
		UpdateResourceMeta(
			util.WithLabel(mesh_proto.ResourceOriginLabel, string(mesh_proto.GlobalResourceOrigin)),
			util.WithoutLabelPrefixes(cfg.Multizone.Global.KDS.Labels.SkipPrefixes...),
		),
		reconcile_v2.If(
			reconcile_v2.And(
				reconcile_v2.TypeIs(system.GlobalSecretType),
				reconcile_v2.NameHasPrefix(system.ZoneTokenSigningKeyPrefix),
			),
			MapZoneTokenSigningKeyGlobalToPublicKey),
		reconcile_v2.If(
			reconcile_v2.IsKubernetes(cfg.Store.Type),
			RemoveK8sSystemNamespaceSuffixMapper(cfg.Store.Kubernetes.SystemNamespace)),
		reconcile_v2.If(
			// we don't want status field from global to be synced to the zone
			reconcile_v2.HasStatus,
			RemoveStatus()),
		reconcile_v2.If(func(resource core_model.Resource) bool {
			// There's a handful of resource types for which we keep the name unchanged
			return !resource.Descriptor().SkipKDSHash
		}, HashSuffixMapper(true, mesh_proto.ZoneTag, mesh_proto.KubeNamespaceTag)),
	}

	zoneMappers := []reconcile_v2.ResourceMapper{
		UpdateResourceMeta(
			util.WithLabel(mesh_proto.ResourceOriginLabel, string(mesh_proto.ZoneResourceOrigin)),
			util.WithLabel(mesh_proto.ZoneTag, cfg.Multizone.Zone.Name),
			util.WithoutLabel(mesh_proto.DeletionGracePeriodStartedLabel),
			util.If(util.IsKubernetes(cfg.Store.Type), util.PopulateNamespaceLabelFromNameExtension()),
			util.WithoutLabelPrefixes(cfg.Multizone.Zone.KDS.Labels.SkipPrefixes...),
		),
		MapInsightResourcesZeroGeneration,
		reconcile_v2.If(
			reconcile_v2.IsKubernetes(cfg.Store.Type),
			RemoveK8sSystemNamespaceSuffixMapper(cfg.Store.Kubernetes.SystemNamespace)),
		HashSuffixMapper(false, mesh_proto.ZoneTag, mesh_proto.KubeNamespaceTag),
	}
	ctx = metadata.AppendToOutgoingContext(ctx, VersionHeader, version.Build.Version)

	reg := registry.Global()
	return &Context{
		ZoneClientCtx:     ctx,
		TypesSentByZone:   reg.ObjectTypes(core_model.SentFromZoneToGlobal()),
		TypesSentByGlobal: reg.ObjectTypes(core_model.SentFromGlobalToZone()),
		GlobalProvidedFilter: CompositeResourceFilters(
			GlobalProvidedFilter(manager),
			SkipUnsupportedHostnameGenerator,
		),
		ZoneProvidedFilter:       ZoneProvidedFilter,
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
		system.ZoneTokenSigningKeyPrefix,
		system.ZoneTokenSigningPublicKeyPrefix,
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

func GlobalProvidedFilter(rm manager.ResourceManager) reconcile_v2.ResourceFilter {
	return func(ctx context.Context, zoneName string, features kds.Features, r core_model.Resource) bool {
		// There's explicit flag to disable KDS for a resource
		if r.Descriptor().HasKDSDisabled(zoneName, r.GetMeta().GetLabels()) {
			return false
		}
		// for Config, Secret and GlobalSecret there are resources that are internal to each CP
		// we filter them out to not send them to other CPs
		// We could instead leverage the KDSDisabled label, but this would be complex for backward compatibility
		switch r.Descriptor().Name {
		case system.ConfigType:
			// We only sync specific entries for configs.
			_, exists := KDSSyncedConfigs[r.GetMeta().GetName()]
			return exists
		case system.GlobalSecretType:
			if slices.Contains([]string{system.EnvoyAdminCA, system.AdminUserToken, system.InterCpCA, system.UserTokenRevocations}, r.GetMeta().GetName()) {
				return false
			}
			if strings.HasPrefix(r.GetMeta().GetName(), system.UserTokenSigningKeyPrefix) {
				return false
			}
		}

		isGlobal := core_model.IsLocallyOriginated(config_core.Global, r.GetMeta().GetLabels())

		switch {
		case isGlobal && r.Descriptor().KDSFlags.Has(core_model.GlobalToZonesFlag):
			if r.Descriptor().IsPluginOriginated && r.Descriptor().IsPolicy {
				policy := r.GetSpec().(core_model.Policy)
				if policy.GetTargetRef().UsesSyntacticSugar && !features.HasFeature(kds.FeatureOptionalTopLevelTargetRef) {
					return false
				}
			}
			return true
		case !isGlobal && r.Descriptor().KDSFlags.Has(core_model.SyncedAcrossZonesFlag):
			if r.Descriptor().IsPluginOriginated && r.Descriptor().IsPolicy {
				if !features.HasFeature(kds.FeatureProducerPolicyFlow) {
					return false
				}
				policy := r.GetSpec().(core_model.Policy)
				if policy.GetTargetRef().UsesSyntacticSugar && !features.HasFeature(kds.FeatureOptionalTopLevelTargetRef) {
					return false
				}
				// if declared role is not 'producer' then no syncing
				if core_model.PolicyRole(r.GetMeta()) != mesh_proto.ProducerPolicyRole {
					return false
				}
				// otherwise we're testing the role in Global CP in case Zone had the validation webhook turned off
				role, err := core_model.ComputePolicyRole(policy, core_model.NewNamespace(r.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag], false))
				if err != nil {
					ri := kri.From(r)
					log.V(1).Info(err.Error(), "name", ri.Name, "mesh", ri.Mesh, "zone", ri.Zone, "namespace", ri.Namespace)
					return false
				}
				// if the actual role is not 'producer' then no syncing
				if role != mesh_proto.ProducerPolicyRole {
					return false
				}
				if policy.GetTargetRef().Kind == common_api.MeshSubset {
					// if top-level targetRef has 'kuma.io/zone' then we can sync it only to required zone
					if targetZone, ok := pointer.Deref(policy.GetTargetRef().Tags)[mesh_proto.ZoneTag]; ok && targetZone != zoneName {
						return false
					}
				}
			}

			// TODO (Icarus9913): replace the function by model.ZoneOfResource(r)
			// Reference: https://github.com/kumahq/kuma/issues/10952
			zoneTag := util.ZoneTag(r)

			// don't need to sync resource to the zone where resource is originating from
			if zoneName == zoneTag {
				return false
			}

			zone := system.NewZoneResource()
			if err := rm.Get(ctx, zone, store.GetByKey(zoneTag, core_model.NoMesh)); err != nil {
				if !errors.Is(err, context.Canceled) {
					log.Error(err, "failed to get zone", "zone", zoneTag)
				}
				// since there is no explicit 'enabled: false' then we don't
				// make any strong decisions which might affect connectivity
				return true
			}
			return zone.Spec.IsEnabled()
		}

		return false
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

// ZoneProvidedFilter filters resources that should be sent from Zone to Global
// a resource is sent from zone to global only if it was created in this zone.
func ZoneProvidedFilter(_ context.Context, localZone string, _ kds.Features, r core_model.Resource) bool {
	if r.Descriptor().HasKDSDisabled("", r.GetMeta().GetLabels()) {
		return false
	}
	return core_model.IsLocallyOriginated(config_core.Zone, r.GetMeta().GetLabels())
}
