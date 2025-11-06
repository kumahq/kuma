package types

import (
	tproxy_dp "github.com/kumahq/kuma/v2/pkg/transparentproxy/config/dataplane"
)

type BootstrapRequest struct {
	Mesh               string  `json:"mesh"`
	Name               string  `json:"name"`
	ProxyType          string  `json:"proxyType"`
	DataplaneToken     string  `json:"dataplaneToken,omitempty"`
	DataplaneTokenPath string  `json:"dataplaneTokenPath,omitempty"`
	DataplaneResource  string  `json:"dataplaneResource,omitempty"`
	Host               string  `json:"-"`
	Version            Version `json:"version"`
	// CaCert is a PEM-encoded CA cert that DP uses to verify CP
	CaCert          string            `json:"caCert"`
	DynamicMetadata map[string]string `json:"dynamicMetadata"`
	DNSPort         uint32            `json:"dnsPort,omitempty"`
	ReadinessPort   uint32            `json:"readinessPort,omitempty"`
	// AppProbeProxyEnabled controls whether the per-pod HTTP probe proxy is enabled.
	//
	// IMPORTANT: Backward compatibility trap
	//
	// The JSON tag for this field is intentionally "appProbeProxyDisabled".
	// The mismatch between the field name and the JSON wire key is historical
	// and has been part of the CP↔DP contract since v2.9.0.
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
	AppProbeProxyEnabled bool                       `json:"appProbeProxyDisabled,omitempty"`
	OperatingSystem      string                     `json:"operatingSystem"`
	Features             []string                   `json:"features"`
	Resources            ProxyResources             `json:"resources"`
	Workdir              string                     `json:"workdir"`
	MetricsResources     MetricsResources           `json:"metricsResources"`
	SystemCaPath         string                     `json:"systemCaPath"`
	TransparentProxy     *tproxy_dp.DataplaneConfig `json:"dataplaneConfig,omitempty"`
	IPv6Enabled          bool                       `json:"ipv6Enabled"`
}

type Version struct {
	KumaDp KumaDpVersion `json:"kumaDp"`
	Envoy  EnvoyVersion  `json:"envoy"`
}

type MetricsResources struct {
	CertPath string `json:"certPath"`
	KeyPath  string `json:"keyPath"`
}

type KumaDpVersion struct {
	Version   string `json:"version"`
	GitTag    string `json:"gitTag"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
}

type EnvoyVersion struct {
	Version          string `json:"version"`
	Build            string `json:"build"`
	KumaDpCompatible bool   `json:"kumaDpCompatible"`
}

// ProxyResources contains information about what resources this proxy has
// available
type ProxyResources struct {
	MaxHeapSizeBytes uint64 `json:"maxHeapSizeBytes"`
}
