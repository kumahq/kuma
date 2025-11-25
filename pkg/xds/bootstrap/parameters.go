package bootstrap

import (
	"time"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	xds_types "github.com/kumahq/kuma/v2/pkg/core/xds/types"
	tproxy_config "github.com/kumahq/kuma/v2/pkg/transparentproxy/config/dataplane"
	"github.com/kumahq/kuma/v2/pkg/xds/bootstrap/types"
)

type KumaDpBootstrap struct {
	AggregateMetricsConfig []AggregateMetricsConfig
	NetworkingConfig       NetworkingConfig
}

type NetworkingConfig struct {
	CorefileTemplate []byte
	Address          string
}

type AggregateMetricsConfig struct {
	Name    string
	Path    string
	Address string
	Port    uint32
}

type configParameters struct {
	Id            string
	Service       string
	AdminAddress  string
	AdminPort     uint32
	ReadinessPort uint32
	// AppProbeProxyEnabled controls whether the per-pod HTTP probe proxy is enabled.
	//
	// IMPORTANT: Backward compatibility trap
	//
	// The BootstrapRequest JSON tag for this setting is intentionally
	// "appProbeProxyDisabled". The mismatch between the field name and the JSON
	// wire key is historical and has been part of the CP↔DP contract since v2.9.0.
	//
	// Why we cannot "fix" this by renaming the field to AppProbeProxyDisabled
	// or by changing the JSON tag:
	//
	// - Mixed-version upgrades are supported. During rolling upgrades there can
	//   be version skew between CP and DP.
	// - If we flip the field name or the JSON tag to match the semantics, the
	//   interpretation in version-skew combinations will invert:
	//     * New DP -> Old CP:
	//         Old CP deserializes "appProbeProxyDisabled: true" into its
	//         AppProbeProxyEnabled = true and will treat the proxy as enabled,
	//         which is the opposite of what the new DP intended.
	//     * Old DP -> New CP:
	//         Old DPs still send the "disabled" wire key. A new CP that expects
	//         a renamed field would read the opposite meaning, again flipping
	//         behavior.
	// - Either case can break probes during upgrades. This violates our
	//   backward-compatibility expectations for rolling upgrades.
	//
	// Possible alternatives like adding a second field, precedence rules, or
	// versioning the bootstrap endpoint would add long-lived complexity and
	// still risk behavior flips mid-upgrade.
	//
	// Therefore, DO NOT rename this field, DO NOT change the JSON tag, and DO
	// NOT invert its meaning. Any attempt to "correct" it will break existing
	// data planes during upgrades.
	//
	// Context: https://github.com/kumahq/kuma/issues/13885
	AppProbeProxyEnabled bool
	AdminAccessLogPath   string
	XdsHost              string
	XdsPort              uint32
	XdsConnectTimeout    time.Duration
	Workdir              string
	MetricsCertPath      string
	MetricsKeyPath       string
	DataplaneToken       string
	DataplaneTokenPath   string
	DataplaneResource    string
	CertBytes            []byte
	Version              *mesh_proto.Version
	HdsEnabled           bool
	DynamicMetadata      map[string]string
	DNSPort              uint32
	ProxyType            string
	Features             xds_types.Features
	IsGatewayDataplane   bool
	Resources            types.ProxyResources
	SystemCaPath         string
	TransparentProxy     *tproxy_config.DataplaneConfig
	IPv6Enabled          bool
	SpireSocketPath      string
}
