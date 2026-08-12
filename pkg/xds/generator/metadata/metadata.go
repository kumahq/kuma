// Package metadata provides lightweight, import-cycle-safe constants shared by
// multiple components (generators, plugins, controllers, hooks, etc.).
// Keeping per-feature constants in a tiny leaf package helps avoid pulling
// heavy transitive dependencies across the build graph and keeps ownership clear
package metadata

import . "github.com/kumahq/kuma/v3/pkg/core/xds/origin"

const (
	// OriginAdmin is the origin for resources produced by the admin proxy/generator
	OriginAdmin Origin = "admin"

	// OriginDirectAccess is the origin for resources produced by the direct-access proxy generator
	OriginDirectAccess Origin = "direct-access"

	// OriginEgress is the origin for resources associated with the egress dataplane/proxy
	OriginEgress Origin = "egress"

	// OriginInbound is the origin for inbound listeners, clusters, and related resources
	OriginInbound Origin = "inbound"

	// OriginIngress is the origin for resources associated with the ingress dataplane/proxy
	OriginIngress Origin = "ingress"

	// OriginOutbound is the origin for outbound listeners, clusters, and related resources
	OriginOutbound Origin = "outbound"

	// OriginPrometheus is the origin for resources produced by the Prometheus endpoint generator
	OriginPrometheus Origin = "prometheus"

	// OriginProxyTemplateModifications is the origin for resources created by MeshProxyPatch.
	// The value is part of the user-facing API (MeshProxyPatch `match.origin`) and must not change.
	OriginProxyTemplateModifications Origin = "proxy-template-modifications"

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

const DirectAccessClusterName = "direct_access"
