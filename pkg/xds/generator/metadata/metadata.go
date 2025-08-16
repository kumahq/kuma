// Package metadata provides lightweight, import-cycle-safe constants shared by
// multiple components (generators, plugins, controllers, hooks, etc.).
// Keeping per-feature constants in a tiny leaf package helps avoid pulling
// heavy transitive dependencies across the build graph and keeps ownership clear
package metadata

import . "github.com/kumahq/kuma/pkg/core/xds/origin"

const (
	// OriginAdmin is the origin for resources produced by the admin proxy/generator
	OriginAdmin Origin = "admin"

	// OriginDirectAccess is the origin for resources produced by the direct-access proxy generator
	OriginDirectAccess Origin = "direct-access"

	// OriginDNS is the origin for resources produced by the DNS generator
	OriginDNS Origin = "dns"

	// OriginEgress is the origin for resources associated with the egress dataplane/proxy
	OriginEgress Origin = "egress"

	// OriginInbound is the origin for inbound listeners, clusters, and related resources
	OriginInbound Origin = "inbound"

	// OriginIngress is the origin for resources associated with the ingress dataplane/proxy
	OriginIngress Origin = "ingress"

	// OriginOutbound is the origin for outbound listeners, clusters, and related resources
	OriginOutbound Origin = "outbound"

	// OriginProbe is the origin for resources produced by the probe/health-check generator
	OriginProbe Origin = "probe"

	// OriginPrometheus is the origin for resources produced by the Prometheus endpoint generator
	OriginPrometheus Origin = "prometheus"

	// OriginProxyTemplateModifications is the origin for resources created by ProxyTemplate modifications
	OriginProxyTemplateModifications Origin = "proxy-template-modifications"

	// OriginProxyTemplateRaw is the origin for resources created by raw ProxyTemplate snippets
	OriginProxyTemplateRaw Origin = "proxy-template-raw"

	// OriginSecrets is the origin for resources produced by the secrets generator
	OriginSecrets Origin = "secrets"

	// OriginTracing is the origin for resources produced by the tracing proxy/generator
	OriginTracing Origin = "tracing"

	// OriginTransparent is the origin for resources produced by the transparent proxy generator
	OriginTransparent Origin = "transparent"
)

const (
	TransparentOutboundNameIPv4  = "outbound:passthrough:ipv4"
	TransparentOutboundNameIPv6  = "outbound:passthrough:ipv6"
	TransparentInboundNameIPv4   = "inbound:passthrough:ipv4"
	TransparentInboundNameIPv6   = "inbound:passthrough:ipv6"
	TransparentInPassThroughIPv4 = "127.0.0.6"
	TransparentInPassThroughIPv6 = "::6"
	TransparentAllIPv4           = "0.0.0.0"
	TransparentAllIPv6           = "::"
)

const (
	ProxyTemplateProfileEgressProxy  = "egress-proxy"
	ProxyTemplateProfileIngressProxy = "ingress-proxy"
)

const (
	ProbeListenerName    = "probe:listener"
	ProbeRouteConfigName = "probe:route_configuration"
)

const DirectAccessClusterName = "direct_access"
