package server

import (
	"github.com/pkg/errors"

	config_store "github.com/kumahq/kuma/pkg/config/core/resources/store"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/providers"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/secrets"
	v3 "github.com/kumahq/kuma/pkg/xds/server/v3"
)

var (
	// HashMeshExcludedResources defines Mesh-scoped resources that are not used in XDS therefore when counting hash mesh we can skip them
	HashMeshExcludedResources = map[core_model.ResourceType]bool{
		core_mesh.DataplaneInsightType:  true,
		core_mesh.DataplaneOverviewType: true,
	}
	HashMeshIncludedGlobalResources = map[core_model.ResourceType]bool{
		core_system.ConfigType:       true,
		core_system.GlobalSecretType: true,
		core_mesh.ZoneIngressType:    true,
		core_mesh.ZoneEgressType:     true,
		core_mesh.MeshType:           true,
	}
)

func MeshResourceTypes() []core_model.ResourceType {
	types := []core_model.ResourceType{}
	for _, desc := range registry.Global().ObjectDescriptors() {
		if desc.Scope == core_model.ScopeMesh && !HashMeshExcludedResources[desc.Name] {
			types = append(types, desc.Name)
		}
	}
	for typ := range HashMeshIncludedGlobalResources {
		types = append(types, typ)
	}
	return types
}

func RegisterXDS(rt core_runtime.Runtime) error {
	// Build common dependencies for V2 and V3 servers.
	// We want to have same metrics (we cannot register one metric twice) and same caches for both V2 and V3.
	statsCallbacks, err := util_xds.NewStatsCallbacks(rt.Metrics(), "xds", util_xds.NoopVersionExtractor)
	if err != nil {
		return err
	}
	claCache, err := cla.NewCache(rt.Config().Store.Cache.ExpirationTime.Duration, rt.Metrics())
	if err != nil {
		return err
	}

	idProvider, err := secrets.NewIdentityProvider(rt.CaManagers(), rt.Metrics())
	if err != nil {
		return err
	}

	secrets, err := secrets.NewSecrets(
		rt.CAProvider(),
		idProvider,
		rt.Metrics(),
	)
	if err != nil {
		return err
	}

	systemNamespace := ""
	if rt.Config().Store.Type == config_store.KubernetesStore {
		systemNamespace = rt.Config().Store.Kubernetes.SystemNamespace
	}
	envoyCpCtx := &xds_context.ControlPlaneContext{
		CLACache:        claCache,
		Secrets:         secrets,
		IdentityManager: providers.NewIdentityProviderManager(rt.IdentityProviders(), rt.EventBus(), rt.Config().Multizone.Zone.Name),
		Zone:            rt.Config().Multizone.Zone.Name,
		SystemNamespace: systemNamespace,
	}

	if err := v3.RegisterXDS(statsCallbacks, rt.XDS().Metrics, envoyCpCtx, rt); err != nil {
		return errors.Wrap(err, "could not register V3 XDS")
	}
	return nil
}
