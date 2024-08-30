package kumadp

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config"
	config_types "github.com/kumahq/kuma/pkg/config/types"
)

var DefaultConfig = func() Config {
	return Config{
		ControlPlane: ControlPlane{
			URL: "https://localhost:5678",
			Retry: CpRetry{
				Backoff:     config_types.Duration{Duration: 3 * time.Second},
				MaxDuration: config_types.Duration{Duration: 5 * time.Minute}, // this value can be fairy long since what will happen when there is a connection error is that the Dataplane will be restarted (by process manager like systemd/K8S etc.) and will try to connect again.
			},
		},
		Dataplane: Dataplane{
			Mesh:          "",
			Name:          "", // Dataplane name must be set explicitly
			DrainTime:     config_types.Duration{Duration: 30 * time.Second},
			ProxyType:     "dataplane",
			ReadinessPort: 9902,
		},
		DataplaneRuntime: DataplaneRuntime{
			BinaryPath: "envoy",
			ConfigDir:  "", // if left empty, a temporary directory will be generated automatically
			DynamicConfiguration: DynamicConfiguration{
				RefreshInterval: config_types.Duration{Duration: 10 * time.Second},
			},
		},
		DNS: DNS{
			Enabled:                   true,
			CoreDNSPort:               15053,
			EnvoyDNSPort:              15054,
			CoreDNSBinaryPath:         "coredns",
			CoreDNSConfigTemplatePath: "",
			ConfigDir:                 "", // if left empty, a temporary directory will be generated automatically
			PrometheusPort:            19153,
			CoreDNSLogging:            false,
		},
	}
}

// Config defines configuration of the Kuma Dataplane Manager.
type Config struct {
	// ControlPlane defines coordinates of the Kuma Control Plane.
	ControlPlane ControlPlane `json:"controlPlane,omitempty"`
	// Dataplane defines bootstrap configuration of the dataplane (Envoy).
	Dataplane Dataplane `json:"dataplane,omitempty"`
	// DataplaneRuntime defines the context in which dataplane (Envoy) runs.
	DataplaneRuntime DataplaneRuntime `json:"dataplaneRuntime,omitempty"`
	// DNS defines a configuration for builtin DNS in Kuma DP
	DNS DNS `json:"dns,omitempty"`
}

func (c *Config) Sanitize() {
	c.ControlPlane.Sanitize()
	c.Dataplane.Sanitize()
	c.DataplaneRuntime.Sanitize()
	c.DNS.Sanitize()
}

func (c *Config) PostProcess() error {
	return multierr.Combine(
		c.ControlPlane.PostProcess(),
		c.Dataplane.PostProcess(),
		c.DataplaneRuntime.PostProcess(),
		c.DNS.PostProcess(),
	)
}

// ControlPlane defines coordinates of the Control Plane.
type ControlPlane struct {
	// URL defines the address of Control Plane DP server.
	URL string `json:"url,omitempty" envconfig:"kuma_control_plane_url"`
	// Retry settings for Control Plane communication
	Retry CpRetry `json:"retry,omitempty"`
	// CaCert defines Certificate Authority that will be used to verify connection to the Control Plane. It takes precedence over CaCertFile.
	CaCert string `json:"caCert" envconfig:"kuma_control_plane_ca_cert"`
	// CaCertFile defines a file for Certificate Authority that will be used to verify connection to the Control Plane.
	CaCertFile string `json:"caCertFile" envconfig:"kuma_control_plane_ca_cert_file"`
}

type ApiServer struct {
	config.BaseConfig

	// Address defines the address of Control Plane API server.
	URL string `json:"url,omitempty" envconfig:"kuma_control_plane_api_server_url"`
	// Retry settings for API Server
	Retry CpRetry `json:"retry,omitempty"`
}

type CpRetry struct {
	config.BaseConfig

	// Duration to wait between retries
	Backoff config_types.Duration `json:"backoff,omitempty" envconfig:"kuma_control_plane_retry_backoff"`
	// Max duration for retries (this is not exact time for execution, the check is done between retries)
	MaxDuration config_types.Duration `json:"maxDuration,omitempty" envconfig:"kuma_control_plane_retry_max_duration"`
}

func (a *CpRetry) Validate() error {
	if a.Backoff.Duration <= 0 {
		return errors.New(".Backoff must be a positive duration")
	}
	if a.MaxDuration.Duration <= 0 {
		return errors.New(".MaxDuration must be a positive duration")
	}
	return nil
}

var _ config.Config = &CpRetry{}

// Dataplane defines bootstrap configuration of the dataplane (Envoy).
type Dataplane struct {
	config.BaseConfig

	// Mesh name.
	Mesh string `json:"mesh,omitempty" envconfig:"kuma_dataplane_mesh"`
	// Dataplane name.
	Name string `json:"name,omitempty" envconfig:"kuma_dataplane_name"`
	// ProxyType defines mode which should be used, supported values: 'dataplane', 'ingress', 'egress'
	ProxyType string `json:"proxyType,omitempty" envconfig:"kuma_dataplane_proxy_type"`
	// Drain time for listeners.
	DrainTime config_types.Duration `json:"drainTime,omitempty" envconfig:"kuma_dataplane_drain_time"`
	// Port that exposes kuma-dp readiness status on localhost, set this value to 0 to provide readiness by "/ready" endpoint from Envoy adminAPI
	ReadinessPort uint32 `json:"readinessPort,omitempty" envconfig:"kuma_readiness_port"`
}

func (d *Dataplane) PostProcess() error {
	if err := validateMeshOrName(".Name", d.Name); err != nil {
		podName, ok := os.LookupEnv("POD_NAME")
		if !ok {
			return nil
		}

		podNamespace, ok := os.LookupEnv("POD_NAMESPACE")
		if !ok {
			return nil
		}

		d.Name = fmt.Sprintf("%s.%s", podName, podNamespace)

		if err := validateMeshOrName(".Name", d.Name); err != nil {
			return errors.Wrap(err, "Dataplane configuration post processing failed")
		}
	}

	return nil
}

func (d *Dataplane) IsZoneProxy() bool {
	return d.ProxyType == string(mesh_proto.IngressProxyType) ||
		d.ProxyType == string(mesh_proto.EgressProxyType)
}

func validateMeshOrName[V ~string](typ string, value V) error {
	if value == "" {
		return errors.Errorf("%s must be non-empty", typ)
	}

	if strings.ContainsAny(string(value), "$(){}") {
		return errors.Errorf("%s %+q contains invalid characters", typ, value)
	}

	return nil
}

// DataplaneRuntime defines the context in which dataplane (Envoy) runs.
type DataplaneRuntime struct {
	config.BaseConfig

	// Path to Envoy binary.
	BinaryPath string `json:"binaryPath,omitempty" envconfig:"kuma_dataplane_runtime_binary_path"`
	// Dir to store auto-generated Envoy bootstrap config in.
	ConfigDir string `json:"configDir,omitempty" envconfig:"kuma_dataplane_runtime_config_dir"`
	// Concurrency specifies how to generate the Envoy concurrency flag.
	Concurrency uint32 `json:"concurrency,omitempty" envconfig:"kuma_dataplane_runtime_concurrency"`
	// Path to a file with dataplane token (use 'kumactl generate dataplane-token' to get one)
	TokenPath string `json:"dataplaneTokenPath,omitempty" envconfig:"kuma_dataplane_runtime_token_path"`
	// Token is dataplane token's value provided directly, will be stored to a temporary file before applying
	Token string `json:"dataplaneToken,omitempty" envconfig:"kuma_dataplane_runtime_token"`
	// Resource is a Dataplane resource that will be applied on Kuma CP
	Resource string `json:"resource,omitempty" envconfig:"kuma_dataplane_runtime_resource"`
	// ResourcePath is a path to Dataplane resource that will be applied on Kuma CP
	ResourcePath string `json:"resourcePath,omitempty" envconfig:"kuma_dataplane_runtime_resource_path"`
	// ResourceVars are the StringToString values that can fill the Resource template
	ResourceVars map[string]string `json:"resourceVars,omitempty"`
	// EnvoyLogLevel is a level on which Envoy will log.
	// Available values are: [trace][debug][info][warning|warn][error][critical][off]
	// By default it inherits Kuma DP logging level.
	EnvoyLogLevel string `json:"envoyLogLevel,omitempty" envconfig:"kuma_dataplane_runtime_envoy_log_level"`
	// EnvoyComponentLogLevel configures Envoy's --component-log-level and uses
	// the exact same syntax: https://www.envoyproxy.io/docs/envoy/latest/operations/cli#cmdoption-component-log-level
	EnvoyComponentLogLevel string `json:"envoyComponentLogLevel,omitempty" envconfig:"kuma_dataplane_runtime_envoy_component_log_level"`
	// Resources defines the resources for this proxy.
	Resources DataplaneResources `json:"resources,omitempty"`
	// SocketDir dir to store socket used between Envoy and the dp process
	SocketDir string `json:"socketDir,omitempty" envconfig:"kuma_dataplane_runtime_socket_dir"`
	// Metrics defines properties of metrics
	Metrics Metrics `json:"metrics,omitempty"`
	// DynamicConfiguration defines properties of dataplane dynamic configuration
	DynamicConfiguration DynamicConfiguration `json:"dynamicConfiguration" envconfig:"kuma_dataplane_runtime_dynamic_configuration"`
	// SystemCaPath defines path of system provided Ca
	SystemCaPath string `json:"systemCaPath,omitempty" envconfig:"kuma_dataplane_runtime_dynamic_system_ca_path"`
}

type Metrics struct {
	// CertPath path to the certificate for metrics listener
	CertPath string `json:"metricsCertPath,omitempty" envconfig:"kuma_dataplane_runtime_metrics_cert_path"`
	// KeyPath path to the key for metrics listener
	KeyPath string `json:"metricsKeyPath,omitempty" envconfig:"kuma_dataplane_runtime_metrics_key_path"`
}

type DynamicConfiguration struct {
	// RefreshInterval defines how often DPP should refresh dynamic config. Default: 10s
	RefreshInterval config_types.Duration `json:"refreshInterval,omitempty" envconfig:"kuma_dataplane_runtime_dynamic_configuration_refresh_interval"`
}

// DataplaneResources defines the resources available to a dataplane proxy.
type DataplaneResources struct {
	MaxMemoryBytes uint64 `json:"maxMemoryBytes,omitempty" envconfig:"kuma_dataplane_resources_max_memory_bytes"`
}

var _ config.Config = &Config{}

func (c *Config) Validate() error {
	var errs error
	if err := c.ControlPlane.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".ControlPlane is not valid"))
	}
	if c.DataplaneRuntime.Resource != "" || c.DataplaneRuntime.ResourcePath != "" {
		if err := c.Dataplane.ValidateForTemplate(); err != nil {
			errs = multierr.Append(errs, errors.Wrapf(err, ".Dataplane is not valid"))
		}
	} else {
		if err := c.Dataplane.Validate(); err != nil {
			errs = multierr.Append(errs, errors.Wrapf(err, ".Dataplane is not valid"))
		}
	}

	if err := c.DataplaneRuntime.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".DataplaneRuntime is not valid"))
	}
	if err := c.DNS.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".DNS is not valid"))
	}
	return errs
}

var _ config.Config = &ControlPlane{}

func (c *ControlPlane) Sanitize() {
	c.Retry.Sanitize()
}

func (c *ControlPlane) PostProcess() error {
	return multierr.Combine(c.Retry.PostProcess())
}

func (c *ControlPlane) Validate() error {
	var errs error
	if _, err := url.Parse(c.URL); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Url is not valid"))
	}
	if err := c.Retry.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Retry is not valid"))
	}
	return errs
}

var _ config.Config = &Dataplane{}

func (d *Dataplane) Validate() error {
	var errs error
	proxyType := mesh_proto.ProxyType(d.ProxyType)
	switch proxyType {
	case mesh_proto.DataplaneProxyType, mesh_proto.IngressProxyType, mesh_proto.EgressProxyType:
	default:
		if err := proxyType.IsValid(); err != nil {
			errs = multierr.Append(errs, errors.Wrap(err, ".ProxyType is not valid"))
		} else {
			// Not all Dataplane types are allowed to be set directly in config.
			errs = multierr.Append(errs, errors.Errorf(".ProxyType %q is not supported", proxyType))
		}
	}

	if d.Mesh == "" && proxyType != mesh_proto.IngressProxyType && proxyType != mesh_proto.EgressProxyType {
		errs = multierr.Append(errs, errors.Errorf(".Mesh must be non-empty"))
	}

	if d.Name == "" {
		errs = multierr.Append(errs, errors.Errorf(".Name must be non-empty"))
	}

	// Notice that d.AdminPort is always valid by design of PortRange
	if d.DrainTime.Duration <= 0 {
		errs = multierr.Append(errs, errors.Errorf(".DrainTime must be positive"))
	}

	if d.ReadinessPort > 65353 {
		return errors.New(".ReadinessPort has to be in [0, 65353] range")
	}

	return errs
}

func (d *Dataplane) ValidateForTemplate() error {
	var errs error
	// Notice that d.AdminPort is always valid by design of PortRange
	if d.DrainTime.Duration <= 0 {
		errs = multierr.Append(errs, errors.Errorf(".DrainTime must be positive"))
	}
	return errs
}

var _ config.Config = &DataplaneRuntime{}

func (d *DataplaneRuntime) Validate() error {
	var errs error
	if d.BinaryPath == "" {
		errs = multierr.Append(errs, errors.Errorf(".BinaryPath must be non-empty"))
	}
	return errs
}

var _ config.Config = &ApiServer{}

func (d *ApiServer) Validate() error {
	var errs error
	if d.URL == "" {
		errs = multierr.Append(errs, errors.Errorf(".URL must be non-empty"))
	}
	if url, err := url.Parse(d.URL); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".URL must be a valid absolute URI"))
	} else if !url.IsAbs() {
		errs = multierr.Append(errs, errors.Errorf(".URL must be a valid absolute URI"))
	}
	if err := d.Retry.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrap(err, ".Retry is not valid"))
	}
	return errs
}

type DNS struct {
	config.BaseConfig

	// If true then builtin DNS functionality is enabled and CoreDNS server is started
	Enabled bool `json:"enabled,omitempty" envconfig:"kuma_dns_enabled"`
	// CoreDNSPort defines a port that handles DNS requests. When transparent proxy is enabled then iptables will redirect DNS traffic to this port.
	CoreDNSPort uint32 `json:"coreDnsPort,omitempty" envconfig:"kuma_dns_core_dns_port"`
	// EnvoyDNSPort defines a port that handles Virtual IP resolving by Envoy. CoreDNS should be configured that it first tries to use this DNS resolver and then the real one.
	EnvoyDNSPort uint32 `json:"envoyDnsPort,omitempty" envconfig:"kuma_dns_envoy_dns_port"`
	// CoreDNSBinaryPath defines a path to CoreDNS binary.
	CoreDNSBinaryPath string `json:"coreDnsBinaryPath,omitempty" envconfig:"kuma_dns_core_dns_binary_path"`
	// CoreDNSConfigTemplatePath defines a path to a CoreDNS config template.
	CoreDNSConfigTemplatePath string `json:"coreDnsConfigTemplatePath,omitempty" envconfig:"kuma_dns_core_dns_config_template_path"`
	// Dir to store auto-generated DNS Server config in.
	ConfigDir string `json:"configDir,omitempty" envconfig:"kuma_dns_config_dir"`
	// PrometheusPort where Prometheus stats will be exposed for the DNS Server
	PrometheusPort uint32 `json:"prometheusPort,omitempty" envconfig:"kuma_dns_prometheus_port"`
	// If true then CoreDNS logging is enabled
	CoreDNSLogging bool `json:"coreDNSLogging,omitempty" envconfig:"kuma_dns_enable_logging"`
}

func (d *DNS) Validate() error {
	if !d.Enabled {
		return nil
	}
	if d.CoreDNSPort > 65353 {
		return errors.New(".CoreDNSPort has to be in [0, 65353] range")
	}
	if d.EnvoyDNSPort > 65353 {
		return errors.New(".EnvoyDNSPort has to be in [0, 65353] range")
	}
	if d.PrometheusPort > 65353 {
		return errors.New(".PrometheusPort has to be in [0, 65353] range")
	}
	if d.CoreDNSBinaryPath == "" {
		return errors.New(".CoreDNSBinaryPath cannot be empty")
	}
	return nil
}
