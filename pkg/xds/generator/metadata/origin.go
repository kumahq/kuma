// This file lives in a dedicated "metadata" subpackage, because the Origin
// type/values are imported by many components (generators, plugins,
// controllers, hooks, etc.). Keeping them in a tiny leaf package avoids
// import cycles and prevents pulling heavy transitive dependencies across
// the build graph. Per-feature constants live in their own metadata
// subpackages to keep ownership clear while keeping dependencies minimal
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
