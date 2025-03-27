package gateway

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"sort"
	"strings"

	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_extensions_transport_sockets_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/core/validators"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	policies_defaults "github.com/kumahq/kuma/pkg/plugins/policies/core/defaults"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	"github.com/kumahq/kuma/pkg/tls"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_secrets "github.com/kumahq/kuma/pkg/xds/envoy/secrets/v3"
	envoy_tls_v3 "github.com/kumahq/kuma/pkg/xds/envoy/tls/v3"
)

// TODO(jpeach) It's a lot to ask operators to tune these defaults,
// and we probably would never do that. However, it would be convenient
// to be able to update them for performance testing and benchmarking,
// so at some point we should consider making these settings available,
// perhaps on the Gateway or on the Dataplane.

// Concurrency defaults.
const DefaultConcurrentStreams = 100

// Window size defaults.
const (
	DefaultInitialStreamWindowSize     = 64 * 1024
	DefaultInitialConnectionWindowSize = 1024 * 1024
)

type keyType string

const (
	keyTypeNone  = keyType("")
	keyTypeECDSA = keyType("ecdsa")
	keyTypeRSA   = keyType("rsa")
)

// HTTPFilterChainGenerator generates a filter chain for a HTTP listener.
type HTTPFilterChainGenerator struct{}

func (g *HTTPFilterChainGenerator) Generate(
	ctx xds_context.Context, info GatewayListenerInfo,
) (
	*core_xds.ResourceSet, []*envoy_listeners.FilterChainBuilder, error,
) {
	log.V(1).Info("generating filter chain", "protocol", "HTTP")

	// HTTP listeners get a single filter chain for all hostnames. So
	// if there's already a filter chain, we have nothing to do.
	chain := newHTTPFilterChain(ctx.Mesh, info)

	chain.Configure(envoy_listeners.HttpDynamicRoute(info.Listener.EnvoyListenerName + ":*"))
	return nil, []*envoy_listeners.FilterChainBuilder{chain}, nil
}

// HTTPSFilterChainGenerator generates a filter chain for an HTTPS listener.
type HTTPSFilterChainGenerator struct{}

func newTLSFilterChain(
	ctx xds_context.Context, info GatewayListenerInfo, listenerHostname GatewayListenerHostname,
) (*core_xds.ResourceSet, *envoy_listeners.FilterChainBuilder, error) {
	log.V(1).Info("generating filter chain", "hostname", listenerHostname.Hostname)

	routeName := listenerHostname.EnvoyRouteName(info.Listener.EnvoyListenerName)
	builder := newHTTPFilterChain(ctx.Mesh, info).Configure(
		envoy_listeners.HttpDynamicRoute(routeName),
	)

	hostResources, err := configureTLS(
		ctx,
		info,
		listenerHostname.TLS,
		[]string{listenerHostname.Hostname},
		builder,
		[]string{"h2", "http/1.1"},
	)
	if err != nil {
		return nil, nil, err
	}

	return hostResources, builder, nil
}

func (g *HTTPSFilterChainGenerator) Generate(
	ctx xds_context.Context, info GatewayListenerInfo,
) (
	*core_xds.ResourceSet, []*envoy_listeners.FilterChainBuilder, error,
) {
	resources := core_xds.NewResourceSet()

	var filterChainBuilders []*envoy_listeners.FilterChainBuilder

	listenerHostnames := info.ListenerHostnames
	if info.Listener.CrossMesh {
		// For cross-mesh, we can only add one listener filter chain as there will not be any (usable) SNI available for filter chain matching
		listenerHostnames = listenerHostnames[:1]
	}
	// In this case we want a single chain for multiple hostnames
	for _, listenerHostname := range listenerHostnames {
		hostResources, builder, err := newTLSFilterChain(ctx, info, listenerHostname)
		if err != nil {
			return nil, nil, err
		}

		filterChainBuilders = append(filterChainBuilders, builder)

		resources.AddSet(hostResources)
	}

	return resources, filterChainBuilders, nil
}

func generateCertificateSecret(
	ctx xds_context.MeshContext,
	hostnames []string,
	secret *system_proto.DataSource,
) (*envoy_extensions_transport_sockets_tls_v3.Secret, error) {
	data, err := ctx.DataSourceLoader.Load(user.Ctx(context.TODO(), user.ControlPlane), ctx.Resource.GetMeta().GetName(), secret)
	if err != nil {
		return nil, err
	}

	tlsSecret, ktype, err := NewServerSecret(data)
	if err != nil {
		return nil, err
	}

	// Generate a name to deterministically identify this secret. We
	// want the same datasource to end up with the same name so that when
	// resources are de-duplicated, we ony have to send the secret once.
	// Since a host can have multiple certificates with
	// different key types, we need to use the key type
	// to disambiguate.
	switch d := secret.GetType().(type) {
	case *system_proto.DataSource_File:
		tlsSecret.Name = names.GetSecretName("cert."+string(ktype), "file", d.File)
	case *system_proto.DataSource_Secret:
		tlsSecret.Name = names.GetSecretName("cert."+string(ktype), "secret", d.Secret)
	case *system_proto.DataSource_Inline:
		tlsSecret.Name = names.GetSecretName("cert."+string(ktype), "inline", names.Join(hostnames...))
	case *system_proto.DataSource_InlineString:
		tlsSecret.Name = names.GetSecretName("cert."+string(ktype), "inlineString", names.Join(hostnames...))
	default:
		return nil, errors.Errorf("unsupported datasource type %T", d)
	}

	return tlsSecret, err
}

func newDownstreamTypedConfig(alpnProtocols []string) *envoy_extensions_transport_sockets_tls_v3.DownstreamTlsContext {
	conf := &envoy_extensions_transport_sockets_tls_v3.DownstreamTlsContext{
		CommonTlsContext: &envoy_extensions_transport_sockets_tls_v3.CommonTlsContext{
			TlsParams:     &envoy_extensions_transport_sockets_tls_v3.TlsParameters{},
			AlpnProtocols: alpnProtocols,
		},
	}

	// TODO(jpeach) add config to set minimum version.
	conf.CommonTlsContext.TlsParams.TlsMinimumProtocolVersion = envoy_extensions_transport_sockets_tls_v3.TlsParameters_TLSv1_2

	// TODO(jpeach) add cipher suite config. The Envoy defaults are pretty good, and enable forward security.
	conf.CommonTlsContext.TlsParams.CipherSuites = nil

	// TODO(jpeach) add config to require a client certificate.
	conf.RequireClientCertificate = util_proto.Bool(false)

	// TODO(jpeach) configure session tickets using SDS.

	return conf
}

func newHTTPFilterChain(ctx xds_context.MeshContext, info GatewayListenerInfo) *envoy_listeners.FilterChainBuilder {
	// A Gateway is a single service across all listeners.
	service := info.Proxy.Dataplane.Spec.GetIdentifyingService()

	builder := envoy_listeners.NewFilterChainBuilder(info.Proxy.APIVersion, envoy_common.AnonymousResource).Configure(
		// Note that even for HTTPS cases, we don't enable client certificate
		// forwarding. This is because this particular configurer will enable
		// forwarding for the client certificate URI, which is OK for SPIFFE-
		// oriented mesh use cases, but unlikely to be appropriate for a
		// general-purpose gateway.
		envoy_listeners.HttpConnectionManager(service, false, info.Proxy.InternalAddresses),
		envoy_listeners.ServerHeader("Kuma Gateway"),
	)

	// Add edge proxy recommendations.
	builder.Configure(
		envoy_listeners.EnablePathNormalization(),
		envoy_listeners.AddFilterChainConfigurer(
			envoy_listeners_v3.HttpConnectionManagerMustConfigureFunc(func(hcm *envoy_hcm.HttpConnectionManager) {
				hcm.UseRemoteAddress = util_proto.Bool(true)

				hcm.RequestHeadersTimeout = util_proto.Duration(policies_defaults.DefaultGatewayRequestHeadersTimeout)
				hcm.StreamIdleTimeout = util_proto.Duration(policies_defaults.DefaultGatewayStreamIdleTimeout)

				hcm.CommonHttpProtocolOptions = &envoy_config_core.HttpProtocolOptions{
					IdleTimeout:                  util_proto.Duration(policies_defaults.DefaultGatewayIdleTimeout),
					HeadersWithUnderscoresAction: envoy_config_core.HttpProtocolOptions_REJECT_REQUEST,
				}

				hcm.Http2ProtocolOptions = &envoy_config_core.Http2ProtocolOptions{
					MaxConcurrentStreams:        util_proto.UInt32(DefaultConcurrentStreams),
					InitialStreamWindowSize:     util_proto.UInt32(DefaultInitialStreamWindowSize),
					InitialConnectionWindowSize: util_proto.UInt32(DefaultInitialConnectionWindowSize),
					AllowConnect:                true,
				}
			}),
		),
	)

	// Tracing and logging have to be configured after the HttpConnectionManager is enabled.
	builder.Configure(
		// Force the ratelimit filter to always be present. This
		// is a no-op unless we later add a per-route configuration.
		envoy_listeners.RateLimit([]*core_mesh.RateLimitResource{nil}),
		envoy_listeners.DefaultCompressorFilter(),
		envoy_listeners.Tracing(
			ctx.GetTracingBackend(info.Proxy.Policies.TrafficTrace),
			service,
			envoy_common.TrafficDirectionUnspecified,
			"",
			true,
		),
		// In mesh proxies, the access log is configured on the outbound
		// listener, which is why we index the Logs slice by destination
		// service name.  A Gateway listener by definition forwards traffic
		// to multiple destinations, so rather than making up some arbitrary
		// rules about which destination service we should accept here, we
		// match the log policy for the generic pass through service. This
		// will be the only policy available for a Dataplane with no outbounds.
		envoy_listeners.HttpAccessLog(
			ctx.Resource.Meta.GetName(),
			envoy_common.TrafficDirectionInbound,
			service,                // Source service is the gateway service.
			mesh_proto.MatchAllTag, // Destination service could be anywhere, depending on the routes.
			ctx.GetLoggingBackend(info.Proxy.Policies.TrafficLogs[core_mesh.PassThroughService]),
			info.Proxy,
		),
	)

	// TODO(jpeach) if proxy protocol is enabled, add the proxy protocol listener filter.

	return builder
}

// NewServerSecret parses a blob that contains one or more PEM object
// to create a single Envoy TLS certificate secret. The resulting secret
// must have exactly one private key, and at least one certificate.
func NewServerSecret(data []byte) (*envoy_extensions_transport_sockets_tls_v3.Secret, keyType, error) {
	var certificates []*pem.Block
	var key *pem.Block
	var ktype keyType

	for i := 0; true; i++ {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}

		switch {
		case strings.HasSuffix(block.Type, "PRIVATE KEY"):
			if key != nil {
				return nil, keyTypeNone, newSecretError(i, "secret contains multiple private keys")
			}

			pkey, err := tls.ParsePrivateKey(block.Bytes)
			if err != nil {
				return nil, keyTypeNone, newSecretError(i, err.Error())
			}

			switch pkey.(type) {
			case *rsa.PrivateKey:
				ktype = keyTypeRSA
			case *ecdsa.PrivateKey:
				ktype = keyTypeECDSA
			default:
				return nil, keyTypeNone, newSecretError(i, fmt.Sprintf("unsupported private key type %T", pkey))
			}

			key = block

		case strings.HasSuffix(block.Type, "CERTIFICATE"):
			certificates = append(certificates, block)

		default:
			return nil, keyTypeNone, newSecretError(i, fmt.Sprintf("unsupported PEM block %q", block.Type))
		}

		data = rest
	}

	if len(certificates) == 0 {
		return nil, keyTypeNone, errors.New("missing server certificate")
	}

	if key == nil {
		return nil, keyTypeNone, errors.New("missing private key")
	}

	return envoy_secrets.NewServerCertificateSecret(key, certificates), ktype, nil
}

func newSecretError(i int, msg string) error {
	var err validators.ValidationError
	err.AddViolationAt(validators.RootedAt("secret").Index(i), msg)
	return err.OrNil()
}

func newTCPFilterChain(
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
	service string,
	clusters []envoy_common.Cluster,
	retryPolicy *core_mesh.RetryResource,
) *envoy_listeners.FilterChainBuilder {
	return envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).Configure(
		envoy_listeners.TcpProxyDeprecated(service, clusters...),
		envoy_listeners.NetworkAccessLog(
			ctx.Mesh.Resource.Meta.GetName(),
			envoy_common.TrafficDirectionInbound,
			service,                // Source service is the gateway service.
			mesh_proto.MatchAllTag, // Destination service could be anywhere, depending on the routes.
			ctx.Mesh.GetLoggingBackend(proxy.Policies.TrafficLogs[core_mesh.PassThroughService]),
			proxy,
		),
		envoy_listeners.MaxConnectAttempts(retryPolicy),
	)
}

// TCPFilterChainGenerator generates a filter chain for a TCP or TLS listener.
type TCPFilterChainGenerator struct{}

func (g *TCPFilterChainGenerator) Generate(
	ctx xds_context.Context, info GatewayListenerInfo,
) (
	*core_xds.ResourceSet, []*envoy_listeners.FilterChainBuilder, error,
) {
	log.V(1).Info("generating filter chain", "protocol", info.Listener.Protocol)

	resources := core_xds.NewResourceSet()

	clustersByHostname := map[string][]envoy_common.Cluster{}
	var allDests []route.Destination

	for _, listenerHostnames := range info.ListenerHostnames {
		for _, host := range listenerHostnames.HostInfos {
			dests := routeDestinations(host.Entries())
			allDests = append(allDests, dests...)

			for _, dest := range dests {
				cluster := envoy_common.NewCluster(
					envoy_common.WithName(dest.Name),
					envoy_common.WithService(dest.Destination[mesh_proto.ServiceTag]),
					envoy_common.WithTags(dest.Destination),
					envoy_common.WithWeight(dest.Weight),
				)
				clustersByHostname[host.Host.Hostname] = append(clustersByHostname[host.Host.Hostname], cluster)
			}
		}
	}
	var allClusters []envoy_common.Cluster
	for _, clusters := range clustersByHostname {
		allClusters = append(allClusters, clusters...)
	}
	sort.Slice(allClusters, func(i, j int) bool { return allClusters[i].Name() < allClusters[j].Name() })

	service := info.Proxy.Dataplane.Spec.GetIdentifyingService()

	// We can only specify retries for the entire filter chain, not per cluster
	var retryPolicy *core_mesh.RetryResource
	if policy := match.BestConnectionPolicyForDestination(allDests, core_mesh.RetryType); policy != nil {
		retryPolicy = policy.(*core_mesh.RetryResource)
	}

	switch info.Listener.Protocol {
	case mesh_proto.MeshGateway_Listener_TLS:
		var filterChains []*envoy_listeners.FilterChainBuilder
		for _, filter := range info.ListenerHostnames {
			clusters := clustersByHostname[filter.Hostname]
			sort.Slice(clusters, func(i, j int) bool { return clusters[i].Name() < clusters[j].Name() })

			builder := newTCPFilterChain(ctx, info.Proxy, service, clusters, retryPolicy)
			tlsResources, err := configureTLS(
				ctx,
				info,
				filter.TLS,
				[]string{filter.Hostname},
				builder,
				nil,
			)
			resources = resources.AddSet(tlsResources)
			if err != nil {
				return nil, nil, err
			}

			filterChains = append(filterChains, builder)
		}
		return resources, filterChains, nil
	default:
		builder := newTCPFilterChain(ctx, info.Proxy, service, allClusters, retryPolicy)
		return resources, []*envoy_listeners.FilterChainBuilder{builder}, nil
	}
}

func configureTLS(
	ctx xds_context.Context,
	listener GatewayListenerInfo,
	tls *mesh_proto.MeshGateway_TLS_Conf,
	hostnames []string,
	builder *envoy_listeners.FilterChainBuilder,
	alpnProtocols []string,
) (
	*core_xds.ResourceSet, error,
) {
	builder.Configure(
		envoy_listeners.MatchTransportProtocol("tls"),
	)

	resources := core_xds.NewResourceSet()

	downstream := newDownstreamTypedConfig(alpnProtocols)

	mode := mesh_proto.MeshGateway_TLS_NONE
	if tls != nil {
		mode = tls.GetMode()
	}

	switch mode {
	case mesh_proto.MeshGateway_TLS_PASSTHROUGH:
		builder.Configure(
			envoy_listeners.MatchServerNames(hostnames...),
		)
		return resources, nil
	case mesh_proto.MeshGateway_TLS_TERMINATE:
		builder.Configure(
			envoy_listeners.MatchServerNames(hostnames...),
		)

		for _, cert := range tls.GetCertificates() {
			secret, err := generateCertificateSecret(ctx.Mesh, hostnames, cert)
			if err != nil {
				return nil, errors.Wrap(err, "failed to generate TLS certificate")
			}

			if resources.Contains(secret.Name, secret) {
				return nil, errors.Errorf("duplicate TLS certificate %q", secret.Name)
			}

			resource := NewResource(secret.Name, secret)

			resources.Add(resource)

			downstream.CommonTlsContext.TlsCertificateSdsSecretConfigs = append(
				downstream.CommonTlsContext.TlsCertificateSdsSecretConfigs, envoy_tls_v3.NewSecretConfigSource(resource.Name),
			)
		}

	case mesh_proto.MeshGateway_TLS_NONE:
		if !listener.Listener.CrossMesh {
			return nil, errors.Errorf("unsupported TLS mode %q", mode)
		}

		// We don't match on the SNI here since it won't be the hostname
		// here but rather an SNI based on the kuma.io/service
		builder.Configure(envoy_listeners.MatchApplicationProtocols("kuma"))

		if !ctx.Mesh.Resource.MTLSEnabled() {
			return nil, errors.New("mTLS must be enabled for crossMesh-enabled MeshGateways")
		}

		var err error
		// We generate a downstream context using the envoy helpers
		// and we don't want to match a SAN because it can be any mesh
		downstream, err = envoy_tls_v3.CreateDownstreamTlsContext(
			listener.Proxy.SecretsTracker.RequestAllInOneCa(),
			listener.Proxy.SecretsTracker.RequestIdentityCert(),
		)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't generate downstream tls context for gateway")
		}

		downstream.CommonTlsContext.GetCombinedValidationContext().DefaultValidationContext.MatchTypedSubjectAltNames = nil
	default:
		return nil, errors.Errorf("unsupported TLS mode %q", tls.GetMode())
	}

	any, err := util_proto.MarshalAnyDeterministic(downstream)
	if err != nil {
		return nil, err
	}

	builder.Configure(
		envoy_listeners.AddFilterChainConfigurer(envoy_listeners_v3.FilterChainMustConfigureFunc(
			func(chain *envoy_listener.FilterChain) {
				chain.TransportSocket = &envoy_config_core.TransportSocket{
					Name: "envoy.transport_sockets.tls",
					ConfigType: &envoy_config_core.TransportSocket_TypedConfig{
						TypedConfig: any,
					},
				}
			}),
		),
	)

	return resources, nil
}
