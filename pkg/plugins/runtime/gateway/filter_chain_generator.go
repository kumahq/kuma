package gateway

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_extensions_transport_sockets_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/tls"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
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
const DefaultInitialStreamWindowSize = 64 * 1024
const DefaultInitialConnectionWindowSize = 1024 * 1024

// Timeout defaults.
const DefaultRequestHeadersTimeout = 500 * time.Millisecond
const DefaultStreamIdleTimeout = 5 * time.Second
const DefaultIdleTimeout = 5 * time.Minute

type keyType string

const keyTypeNone = keyType("")
const keyTypeECDSA = keyType("ecdsa")
const keyTypeRSA = keyType("rsa")

// HTTPFilterChainGenerator generates a filter chain for a HTTP listener.
type HTTPFilterChainGenerator struct{}

func (*HTTPFilterChainGenerator) SupportsProtocol(p mesh_proto.Gateway_Listener_Protocol) bool {
	return p == mesh_proto.Gateway_Listener_HTTP
}

func (*HTTPFilterChainGenerator) GenerateHost(ctx xds_context.Context, info *GatewayResourceInfo) (*core_xds.ResourceSet, error) {
	// HTTP listeners get a single filter chain for all hostnames. So
	// if there's already a filter chain, we have nothing to do.
	if info.Resources.FilterChain != nil {
		log.V(1).Info("updating existing filter chain",
			"hostname", info.Host.Hostname,
		)

		return nil, nil
	}

	log.V(1).Info("generating filter chain",
		"hostname", info.Host.Hostname,
	)

	info.Resources.FilterChain = newFilterChain(ctx, info)
	return nil, nil
}

// HTTPSFilterChainGenerator generates a filter chain for an HTTPS listener.
type HTTPSFilterChainGenerator struct {
	DataSourceLoader datasource.Loader
}

func (*HTTPSFilterChainGenerator) SupportsProtocol(p mesh_proto.Gateway_Listener_Protocol) bool {
	return p == mesh_proto.Gateway_Listener_HTTPS
}

func (g *HTTPSFilterChainGenerator) GenerateHost(ctx xds_context.Context, info *GatewayResourceInfo) (*core_xds.ResourceSet, error) {
	// HTTPS listeners get a filter chain for each hostname. So if
	// there is already a filter chain (from the previous host), we
	// close it, and start a new one.
	if info.Resources.FilterChain != nil {
		info.Resources.Listener.Configure(
			envoy_listeners.FilterChain(info.Resources.FilterChain),
		)
	}

	log.V(1).Info("generating filter chain",
		"hostname", info.Host.Hostname,
	)

	resources := core_xds.NewResourceSet()

	switch info.Host.TLS.GetMode() {
	case mesh_proto.Gateway_TLS_TERMINATE:
		// Note that Envoy 1.184 and earlier will only accept 1 SDS reference.
		for _, cert := range info.Host.TLS.GetCertificates() {
			secret, err := g.generateCertificateSecret(ctx, info, cert)
			if err != nil {
				return nil, errors.Wrap(err, "failed to generate TLS certificate")
			}

			if resources.Contains(secret.Name, secret) {
				return nil, errors.Errorf("duplicate TLS certificate %q", secret.Name)
			}

			resources.Add(NewResource(secret.Name, secret))
		}

	case mesh_proto.Gateway_TLS_PASSTHROUGH:
		// TODO(jpeach) add support for PASSTHROUGH mode.
		return nil, errors.Errorf("unsupported TLS mode %q", info.Host.TLS.GetMode())

	default:
		return nil, errors.Errorf("unsupported TLS mode %q", info.Host.TLS.GetMode())
	}

	builder := newFilterChain(ctx, info)

	builder.Configure(
		envoy_listeners.MatchTransportProtocol("tls"),
		envoy_listeners.MatchServerNames(info.Host.Hostname),
		envoy_listeners.MatchApplicationProtocols("h2", "http/1.1"),
	)

	downstream := newDownstreamTypedConfig()

	// If we generated any secrets, attach their references to the downstream validation context.
	for _, s := range resources.ListOf(envoy_resource.SecretType) {
		downstream.CommonTlsContext.TlsCertificateSdsSecretConfigs = append(
			downstream.CommonTlsContext.TlsCertificateSdsSecretConfigs, envoy_tls_v3.NewSecretConfigSource(s.Name),
		)
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

	// Leave the builder open so that later steps can update the filters.
	info.Resources.FilterChain = builder

	return resources, nil
}

func (g *HTTPSFilterChainGenerator) generateCertificateSecret(
	ctx xds_context.Context,
	info *GatewayResourceInfo,
	secret *system_proto.DataSource,
) (*envoy_extensions_transport_sockets_tls_v3.Secret, error) {
	data, err := g.DataSourceLoader.Load(context.Background(), ctx.Mesh.Resource.GetMeta().GetName(), secret)
	if err != nil {
		return nil, err
	}

	tlsSecret, ktype, err := newServerSecret(data)
	if err != nil {
		return nil, err
	}

	// Generate a name to deterministically identify this secret. We
	// want the same datasource to end up with the same name so that when
	// resources are de-duplicated, we ony have to send the secret once.
	switch d := secret.GetType().(type) {
	case *system_proto.DataSource_File:
		tlsSecret.Name = names.GetSecretName("cert."+string(ktype), "file", d.File)
	case *system_proto.DataSource_Secret:
		tlsSecret.Name = names.GetSecretName("cert."+string(ktype), "secret", d.Secret)
	case *system_proto.DataSource_Inline:
		// Since a host can have multiple certificates with
		// different key types, we need to use the key type
		// to disambiguate when the certificate is provided as
		// inline data.
		tlsSecret.Name = names.GetSecretName("cert."+string(ktype), "inline", info.Host.Hostname)
	default:
		return nil, errors.Errorf("unsupported datasource type %T", d)
	}

	return tlsSecret, err
}

func newDownstreamTypedConfig() *envoy_extensions_transport_sockets_tls_v3.DownstreamTlsContext {
	conf := &envoy_extensions_transport_sockets_tls_v3.DownstreamTlsContext{
		CommonTlsContext: &envoy_extensions_transport_sockets_tls_v3.CommonTlsContext{
			TlsParams:     &envoy_extensions_transport_sockets_tls_v3.TlsParameters{},
			AlpnProtocols: []string{"h2", "http/1.1"},
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

func newFilterChain(ctx xds_context.Context, info *GatewayResourceInfo) *envoy_listeners.FilterChainBuilder {
	// A Gateway is a single service across all listeners.
	service := info.Dataplane.Spec.GetIdentifyingService()

	builder := envoy_listeners.NewFilterChainBuilder(info.Proxy.APIVersion).Configure(
		// Note that even for HTTPS cases, we don't enable client certificate
		// forwarding. This is because this particular configurer will enable
		// forwarding for the client certificate URI, which is OK for SPIFFE-
		// oriented mesh use cases, but unlikely to be appropriate for a
		// general-purpose gateway.
		envoy_listeners.HttpConnectionManager(service, false),
		envoy_listeners.ServerHeader("Kuma Gateway"),
		envoy_listeners.HttpDynamicRoute(info.Listener.ResourceName),
	)

	// Add edge proxy recommendations.
	builder.Configure(
		envoy_listeners.EnablePathNormalization(),
		envoy_listeners.StripHostPort(),
		envoy_listeners.AddFilterChainConfigurer(
			envoy_listeners_v3.HttpConnectionManagerMustConfigureFunc(func(hcm *envoy_hcm.HttpConnectionManager) {
				hcm.RequestHeadersTimeout = util_proto.Duration(DefaultRequestHeadersTimeout)
				hcm.StreamIdleTimeout = util_proto.Duration(DefaultStreamIdleTimeout)

				hcm.CommonHttpProtocolOptions = &envoy_config_core.HttpProtocolOptions{
					IdleTimeout:                  util_proto.Duration(DefaultIdleTimeout),
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
		envoy_listeners.DefaultCompressorFilter(),
		envoy_listeners.Tracing(info.Proxy.Policies.TracingBackend, service),
		// In mesh proxies, the access log is configured on the outbound
		// listener, which is why we index the Logs slice by destination
		// service name.  A Gateway listener by definition forwards traffic
		// to multiple destinations, so rather than making up some arbitrary
		// rules about which destination service we should accept here, we
		// match the log policy for the generic pass through service. This
		// will be the only policy available for a Dataplane with no outbounds.
		envoy_listeners.HttpAccessLog(
			ctx.Mesh.Resource.Meta.GetName(),
			envoy.TrafficDirectionInbound,
			service,                // Source service is the gateway service.
			mesh_proto.MatchAllTag, // Destination service could be anywhere, depending on the routes.
			info.Proxy.Policies.Logs[core_mesh.PassThroughService],
			info.Proxy,
		),
	)

	// TODO(jpeach) if proxy protocol is enabled, add the proxy protocol listener filter.

	return builder
}

// newServerSecret parses a blob that contains one or more PEM object
// to create a single Envoy TLS certificate secret. The resulting secret
// must have exactly one private key, and at least one certificate.
func newServerSecret(data []byte) (*envoy_extensions_transport_sockets_tls_v3.Secret, keyType, error) {
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
