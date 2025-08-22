package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"

	unified_naming "github.com/kumahq/kuma/pkg/core/naming/unified-naming"
	core_system_names "github.com/kumahq/kuma/pkg/core/system_names"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/core/xds/types"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/generator/metadata"
	"github.com/kumahq/kuma/pkg/xds/generator/system_names"
)

var staticEndpointPaths = []*envoy_common.StaticEndpointPath{
	{
		Path:        "/ready",
		RewritePath: "/ready",
	},
}

var staticTlsEndpointPaths = []*envoy_common.StaticEndpointPath{
	{
		Path:        "/ready",
		RewritePath: "/ready",
	},
	{
		Path:        "/",
		RewritePath: "/",
	},
}

// AdminProxyGenerator generates resources to expose some endpoints of Admin API on public interface.
// By default, Admin API is exposed only on loopback interface because of security reasons.
type AdminProxyGenerator struct{}

var adminAddressAllowedValues = map[string]struct{}{
	"127.0.0.1": {},
	"0.0.0.0":   {},
	"::1":       {},
	"::":        {},
	"":          {},
}

func (g AdminProxyGenerator) Generate(ctx context.Context, _ *core_xds.ResourceSet, xdsCtx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	if proxy.Metadata.GetAdminPort() == 0 {
		// It's not possible to export Admin endpoints if Envoy Admin API has not been enabled on that dataplane.
		return nil, nil
	}

	adminPort := proxy.Metadata.GetAdminPort()
	readinessPort := proxy.Metadata.GetReadinessPort()
	if readinessPort == 0 {
		return nil, errors.New("ReadinessPort has to be in (0, 65353] range")
	}
	// TODO(unified-resource-naming): adjust when legacy naming is removed
	unifiedNamingEnabled := unified_naming.Enabled(proxy.Metadata, xdsCtx.Mesh.Resource)
	// We assume that Admin API must be available on a loopback interface (while users
	// can override the default value `127.0.0.1` in the Bootstrap Server section of `kuma-cp` config,
	// the only reasonable alternatives are `::1`, `0.0.0.0` or `::`).
	// In contrast to `AdminPort`, we shouldn't trust `AdminAddress` from the Envoy node metadata
	// since it would allow a malicious user to manipulate that value and use Prometheus endpoint
	// as a gateway to another host.
	envoyAdminClusterName := envoy_names.GetEnvoyAdminClusterName()
	if unifiedNamingEnabled {
		envoyAdminClusterName = system_names.SystemResourceNameEnvoyAdmin
	}

	getNameOrDefault := core_system_names.GetNameOrDefault(unifiedNamingEnabled)
	dppReadinessClusterName := getNameOrDefault(
		system_names.SystemResourceNameReadiness,
		envoy_names.GetDPPReadinessClusterName(),
	)
	adminAddress := proxy.Metadata.GetAdminAddress()
	if _, ok := adminAddressAllowedValues[adminAddress]; !ok {
		var allowedAddresses []string
		for _, address := range util_maps.SortedKeys(adminAddressAllowedValues) {
			allowedAddresses = append(allowedAddresses, fmt.Sprintf(`%q`, address))
		}
		return nil, errors.Errorf("envoy admin cluster is not allowed to have addresses other than %s", strings.Join(allowedAddresses, ", "))
	}
	switch adminAddress {
	case "", "0.0.0.0":
		adminAddress = "127.0.0.1"
	case "::":
		adminAddress = "::1"
	}

	envoyAdminCluster, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, envoyAdminClusterName).
		Configure(envoy_clusters.ProvidedEndpointCluster(
			govalidator.IsIPv6(adminAddress),
			core_xds.Endpoint{Target: adminAddress, Port: adminPort})).
		Configure(envoy_clusters.DefaultTimeout()).
		Build()
	if err != nil {
		return nil, err
	}

	for _, se := range staticEndpointPaths {
		se.ClusterName = dppReadinessClusterName
	}
	for _, se := range staticTlsEndpointPaths {
		switch se.Path {
		case "/ready":
			se.ClusterName = dppReadinessClusterName
		default:
			se.ClusterName = envoyAdminClusterName
		}
	}

	resources := core_xds.NewResourceSet()
	// We bind admin to 127.0.0.1 by default, creating another listener with same address and port will result in error.
	if g.getAddress(proxy) != adminAddress {
		// TODO(unified-resource-naming): adjust when legacy naming is removed
		envoyAdminListenerName := envoy_names.GetAdminListenerName()
		if unifiedNamingEnabled {
			envoyAdminListenerName = system_names.SystemResourceNameEnvoyAdmin
		}
		filterChains := []envoy_listeners.ListenerBuilderOpt{
			envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
				Configure(envoy_listeners.StaticEndpoints(envoyAdminListenerName, staticEndpointPaths)),
			),
		}
		filterChains = append(filterChains, envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
			Configure(envoy_listeners.MatchTransportProtocol("tls")).
			Configure(envoy_listeners.StaticEndpoints(envoyAdminListenerName, staticTlsEndpointPaths)).
			Configure(envoy_listeners.ServerSideStaticMTLS(proxy.EnvoyAdminMTLSCerts)),
		))

		listener, err := envoy_listeners.NewInboundListenerBuilder(proxy.APIVersion, g.getAddress(proxy), adminPort, core_xds.SocketAddressProtocolTCP).
			WithOverwriteName(envoyAdminListenerName).
			Configure(envoy_listeners.TLSInspector()).
			Configure(filterChains...).
			Build()
		if err != nil {
			return nil, err
		}
		resources.Add(&core_xds.Resource{
			Name:     listener.GetName(),
			Origin:   metadata.OriginAdmin,
			Resource: listener,
		})
	}

	resources.Add(&core_xds.Resource{
		Name:     envoyAdminCluster.GetName(),
		Origin:   metadata.OriginAdmin,
		Resource: envoyAdminCluster,
	})

	var xdsEndpoint core_xds.Endpoint
	if proxy.Metadata.HasFeature(types.FeatureReadinessUnixSocket) {
		xdsEndpoint = core_xds.Endpoint{
			UnixDomainPath: core_xds.ReadinessReporterSocketName(proxy.Metadata.WorkDir),
		}
	} else {
		xdsEndpoint = core_xds.Endpoint{
			Target: adminAddress,
			Port:   readinessPort,
		}
	}

	readinessCluster, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, dppReadinessClusterName).
		Configure(envoy_clusters.ProvidedEndpointCluster(govalidator.IsIPv6(adminAddress), xdsEndpoint)).
		Configure(envoy_clusters.DefaultTimeout()).
		Build()
	if err != nil {
		return nil, err
	}

	resources.Add(&core_xds.Resource{
		Name:     readinessCluster.GetName(),
		Origin:   metadata.OriginAdmin,
		Resource: readinessCluster,
	})

	return resources, nil
}

func (g AdminProxyGenerator) getAddress(proxy *core_xds.Proxy) string {
	if proxy.Dataplane != nil {
		return proxy.Dataplane.Spec.GetNetworking().Address
	}

	if proxy.ZoneEgressProxy != nil {
		return proxy.ZoneEgressProxy.ZoneEgressResource.Spec.GetNetworking().GetAddress()
	}

	if proxy.ZoneIngressProxy != nil {
		return proxy.ZoneIngressProxy.ZoneIngressResource.Spec.GetNetworking().GetAddress()
	}

	return ""
}
