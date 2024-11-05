package kuma_cp

import (
	"net"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/config/access"
	api_server "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/config/core/resources/apis"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/config/diagnostics"
	dns_server "github.com/kumahq/kuma/pkg/config/dns-server"
	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	"github.com/kumahq/kuma/pkg/config/eventbus"
	"github.com/kumahq/kuma/pkg/config/intercp"
	"github.com/kumahq/kuma/pkg/config/mads"
	"github.com/kumahq/kuma/pkg/config/multizone"
	"github.com/kumahq/kuma/pkg/config/plugins/policies"
	"github.com/kumahq/kuma/pkg/config/plugins/runtime"
	"github.com/kumahq/kuma/pkg/config/tracing"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/config/xds"
	"github.com/kumahq/kuma/pkg/config/xds/bootstrap"
)

var _ config.Config = &Config{}

var _ config.Config = &Defaults{}

type Defaults struct {
	config.BaseConfig

	// If true, it skips creating the default Mesh
	SkipMeshCreation bool `json:"skipMeshCreation" envconfig:"kuma_defaults_skip_mesh_creation"`
	// If true, it skips creating the default tenant resources
	SkipTenantResources bool `json:"skipTenantResources" envconfig:"kuma_defaults_skip_tenant_resources"`
	// If true, automatically create the default routing (TrafficPermission and TrafficRoute) resources for a new Mesh.
	// These policies are essential for traffic to flow correctly when operating a global control plane with zones running older (<2.6.0) versions of Kuma.
	CreateMeshRoutingResources bool `json:"createMeshRoutingResources" envconfig:"kuma_defaults_create_mesh_routing_resources"`
	// If true, it skips creating default hostname generators
	SkipHostnameGenerators bool `json:"SkipHostnameGenerators" envconfig:"kuma_defaults_skip_hostname_generators"`
}

type Metrics struct {
	config.BaseConfig

	Dataplane    *DataplaneMetrics    `json:"dataplane"`
	Zone         *ZoneMetrics         `json:"zone"`
	Mesh         *MeshMetrics         `json:"mesh"`
	ControlPlane *ControlPlaneMetrics `json:"controlPlane"`
}

func (m *Metrics) Validate() error {
	if err := m.Dataplane.Validate(); err != nil {
		return errors.Wrap(err, "Dataplane validation failed")
	}
	return nil
}

type DataplaneMetrics struct {
	config.BaseConfig

	SubscriptionLimit int                   `json:"subscriptionLimit" envconfig:"kuma_metrics_dataplane_subscription_limit"`
	IdleTimeout       config_types.Duration `json:"idleTimeout" envconfig:"kuma_metrics_dataplane_idle_timeout"`
}

func (d *DataplaneMetrics) Validate() error {
	if d.SubscriptionLimit < 0 {
		return errors.New("SubscriptionLimit should be positive or equal 0")
	}
	return nil
}

type ZoneMetrics struct {
	config.BaseConfig

	SubscriptionLimit int                   `json:"subscriptionLimit" envconfig:"kuma_metrics_zone_subscription_limit"`
	IdleTimeout       config_types.Duration `json:"idleTimeout" envconfig:"kuma_metrics_zone_idle_timeout"`
	// CompactFinishedSubscriptions compacts finished metrics (do not store config and details of KDS exchange).
	CompactFinishedSubscriptions bool `json:"compactFinishedSubscriptions" envconfig:"kuma_metrics_zone_compact_finished_subscriptions"`
}

func (d *ZoneMetrics) Validate() error {
	if d.SubscriptionLimit < 0 {
		return errors.New("SubscriptionLimit should be positive or equal 0")
	}
	return nil
}

type MeshMetrics struct {
	config.BaseConfig

	// Deprecated: use MinResyncInterval instead
	MinResyncTimeout config_types.Duration `json:"minResyncTimeout" envconfig:"kuma_metrics_mesh_min_resync_timeout"`
	// Deprecated: use FullResyncInterval instead
	MaxResyncTimeout config_types.Duration `json:"maxResyncTimeout" envconfig:"kuma_metrics_mesh_max_resync_timeout"`
	// BufferSize the size of the buffer between event creation and processing
	BufferSize int `json:"bufferSize" envconfig:"kuma_metrics_mesh_buffer_size"`
	// MinResyncInterval the minimum time between 2 refresh of insights.
	MinResyncInterval config_types.Duration `json:"minResyncInterval" envconfig:"kuma_metrics_mesh_min_resync_interval"`
	// FullResyncInterval time between triggering a full refresh of all the insights
	FullResyncInterval config_types.Duration `json:"fullResyncInterval" envconfig:"kuma_metrics_mesh_full_resync_interval"`
	// EventProcessors is a number of workers that process metrics events.
	EventProcessors int `json:"eventProcessors" envconfig:"kuma_metrics_mesh_event_processors"`
}

type ControlPlaneMetrics struct {
	// ReportResourcesCount if true will report metrics with the count of resources.
	// Default: true
	ReportResourcesCount bool `json:"reportResourcesCount" envconfig:"kuma_metrics_control_plane_report_resources_count"`
}

func (d *MeshMetrics) Validate() error {
	if d.MinResyncTimeout.Duration != 0 && d.MaxResyncTimeout.Duration <= d.MinResyncTimeout.Duration {
		return errors.New("FullResyncInterval should be greater than MinResyncInterval")
	}
	if d.MinResyncInterval.Duration <= d.FullResyncInterval.Duration {
		return errors.New("FullResyncInterval should be greater than MinResyncInterval")
	}
	return nil
}

type Reports struct {
	// If true then usage stats will be reported
	Enabled bool `json:"enabled" envconfig:"kuma_reports_enabled"`
}

type Config struct {
	// General configuration
	General *GeneralConfig `json:"general,omitempty"`
	// Environment Type, can be either "kubernetes" or "universal"
	Environment core.EnvironmentType `json:"environment,omitempty" envconfig:"kuma_environment"`
	// Mode in which Kuma CP is running. Available values are: "standalone", "global", "zone"
	Mode core.CpMode `json:"mode" envconfig:"kuma_mode"`
	// Resource Store configuration
	Store *store.StoreConfig `json:"store,omitempty"`
	// Configuration of Bootstrap Server, which provides bootstrap config to Dataplanes
	BootstrapServer *bootstrap.BootstrapServerConfig `json:"bootstrapServer,omitempty"`
	// Envoy XDS server configuration
	XdsServer *xds.XdsServerConfig `json:"xdsServer,omitempty"`
	// Monitoring Assignment Discovery Service (MADS) server configuration
	MonitoringAssignmentServer *mads.MonitoringAssignmentServerConfig `json:"monitoringAssignmentServer,omitempty"`
	// API Server configuration
	ApiServer *api_server.ApiServerConfig `json:"apiServer,omitempty"`
	// Environment-specific configuration
	Runtime *runtime.RuntimeConfig `json:"runtime,omitempty"`
	// Default Kuma entities configuration
	Defaults *Defaults `json:"defaults,omitempty"`
	// Metrics configuration
	Metrics *Metrics `json:"metrics,omitempty"`
	// Reports configuration
	Reports *Reports `json:"reports,omitempty"`
	// Multizone Config
	Multizone *multizone.MultizoneConfig `json:"multizone,omitempty"`
	// DNS Server Config
	DNSServer *dns_server.Config `json:"dnsServer,omitempty"`
	// Diagnostics configuration
	Diagnostics *diagnostics.DiagnosticsConfig `json:"diagnostics,omitempty"`
	// Dataplane Server configuration
	DpServer *dp_server.DpServerConfig `json:"dpServer"`
	// Access Control configuration
	Access access.AccessConfig `json:"access"`
	// Configuration of experimental features
	Experimental ExperimentalConfig `json:"experimental"`
	// Proxy holds configuration for proxies
	Proxy xds.Proxy `json:"proxy"`
	// Intercommunication CP configuration
	InterCp intercp.InterCpConfig `json:"interCp"`
	// Tracing
	Tracing tracing.Config `json:"tracing"`
	// EventBus is a configuration of the event bus which is local to one instance of CP.
	EventBus eventbus.Config `json:"eventBus"`
	// Policies is a configuration of plugin policies like MeshAccessLog, MeshTrace etc.
	Policies *policies.Config `json:"policies"`
	// CoreResources holds configuration for generated core resources like MeshService
	CoreResources *apis.Config `json:"coreResources"`
	// IP administration and management config
	IPAM IPAMConfig `json:"ipam"`
	// MeshService holds configuration for features around MeshServices
	MeshService MeshServiceConfig `json:"meshService"`
}

func (c Config) IsFederatedZoneCP() bool {
	return c.Mode == core.Zone && c.Multizone.Zone.GlobalAddress != "" && c.Multizone.Zone.Name != ""
}

func (c Config) IsNonFederatedZoneCP() bool {
	return c.Mode == core.Zone && !c.IsFederatedZoneCP()
}

func (c *Config) Sanitize() {
	c.General.Sanitize()
	c.Store.Sanitize()
	c.BootstrapServer.Sanitize()
	c.XdsServer.Sanitize()
	c.MonitoringAssignmentServer.Sanitize()
	c.ApiServer.Sanitize()
	c.Runtime.Sanitize()
	c.Metrics.Sanitize()
	c.Defaults.Sanitize()
	c.DNSServer.Sanitize()
	c.Multizone.Sanitize()
	c.Diagnostics.Sanitize()
	c.Policies.Sanitize()
}

func (c *Config) PostProcess() error {
	return multierr.Combine(
		c.General.PostProcess(),
		c.Store.PostProcess(),
		c.BootstrapServer.PostProcess(),
		c.XdsServer.PostProcess(),
		c.MonitoringAssignmentServer.PostProcess(),
		c.ApiServer.PostProcess(),
		c.Runtime.PostProcess(),
		c.Metrics.PostProcess(),
		c.Defaults.PostProcess(),
		c.DNSServer.PostProcess(),
		c.Multizone.PostProcess(),
		c.Diagnostics.PostProcess(),
		c.Policies.PostProcess(),
	)
}

var DefaultConfig = func() Config {
	return Config{
		Environment:                core.UniversalEnvironment,
		Mode:                       core.Zone,
		Store:                      store.DefaultStoreConfig(),
		XdsServer:                  xds.DefaultXdsServerConfig(),
		MonitoringAssignmentServer: mads.DefaultMonitoringAssignmentServerConfig(),
		ApiServer:                  api_server.DefaultApiServerConfig(),
		BootstrapServer:            bootstrap.DefaultBootstrapServerConfig(),
		Runtime:                    runtime.DefaultRuntimeConfig(),
		Defaults:                   DefaultDefaultsConfig(),
		Metrics: &Metrics{
			Dataplane: &DataplaneMetrics{
				SubscriptionLimit: 2,
				IdleTimeout:       config_types.Duration{Duration: 5 * time.Minute},
			},
			Zone: &ZoneMetrics{
				SubscriptionLimit: 10,
				IdleTimeout:       config_types.Duration{Duration: 5 * time.Minute},
			},
			Mesh: &MeshMetrics{
				MinResyncInterval:  config_types.Duration{Duration: 1 * time.Second},
				FullResyncInterval: config_types.Duration{Duration: 20 * time.Second},
				BufferSize:         1000,
				EventProcessors:    1,
			},
			ControlPlane: &ControlPlaneMetrics{
				ReportResourcesCount: true,
			},
		},
		Reports: &Reports{
			Enabled: false,
		},
		General:     DefaultGeneralConfig(),
		DNSServer:   dns_server.DefaultDNSServerConfig(),
		Multizone:   multizone.DefaultMultizoneConfig(),
		Diagnostics: diagnostics.DefaultDiagnosticsConfig(),
		DpServer:    dp_server.DefaultDpServerConfig(),
		Access:      access.DefaultAccessConfig(),
		Experimental: ExperimentalConfig{
			KubeOutboundsAsVIPs:             true,
			UseTagFirstVirtualOutboundModel: false,
			IngressTagFilters:               []string{},
			KDSEventBasedWatchdog: ExperimentalKDSEventBasedWatchdog{
				Enabled:            false,
				FlushInterval:      config_types.Duration{Duration: 5 * time.Second},
				FullResyncInterval: config_types.Duration{Duration: 1 * time.Minute},
				DelayFullResync:    false,
			},
			SidecarContainers: false,
		},
		Proxy:         xds.DefaultProxyConfig(),
		InterCp:       intercp.DefaultInterCpConfig(),
		EventBus:      eventbus.Default(),
		Policies:      policies.Default(),
		CoreResources: apis.Default(),
		IPAM: IPAMConfig{
			MeshService: MeshServiceIPAM{
				CIDR: "241.0.0.0/8",
			},
			MeshExternalService: MeshExternalServiceIPAM{
				CIDR: "242.0.0.0/8",
			},
			MeshMultiZoneService: MeshMultiZoneServiceIPAM{
				CIDR: "243.0.0.0/8",
			},
			AllocationInterval: config_types.Duration{Duration: 5 * time.Second},
		},
		MeshService: MeshServiceConfig{
			GenerationInterval:  config_types.Duration{Duration: 2 * time.Second},
			DeletionGracePeriod: config_types.Duration{Duration: 1 * time.Hour},
		},
	}
}

func (c *Config) Validate() error {
	if err := core.ValidateCpMode(c.Mode); err != nil {
		return errors.Wrap(err, "Mode validation failed")
	}
	switch c.Mode {
	case core.Global:
		if err := c.Multizone.Global.Validate(); err != nil {
			return errors.Wrap(err, "Multizone Global validation failed")
		}
	case core.Zone:
		if err := c.Multizone.Zone.Validate(); err != nil {
			return errors.Wrap(err, "Multizone Zone validation failed")
		}
		if err := c.XdsServer.Validate(); err != nil {
			return errors.Wrap(err, "Xds Server validation failed")
		}
		if err := c.BootstrapServer.Validate(); err != nil {
			return errors.Wrap(err, "Bootstrap Server validation failed")
		}
		if err := c.MonitoringAssignmentServer.Validate(); err != nil {
			return errors.Wrap(err, "Monitoring Assignment Server validation failed")
		}
		if c.Environment != core.KubernetesEnvironment && c.Environment != core.UniversalEnvironment {
			return errors.Errorf("Environment should be either %s or %s", core.KubernetesEnvironment, core.UniversalEnvironment)
		}
		if err := c.Runtime.Validate(c.Environment); err != nil {
			return errors.Wrap(err, "Runtime validation failed")
		}
		if err := c.Metrics.Validate(); err != nil {
			return errors.Wrap(err, "Metrics validation failed")
		}
	}
	if err := c.Store.Validate(); err != nil {
		return errors.Wrap(err, "Store validation failed")
	}
	if err := c.ApiServer.Validate(); err != nil {
		return errors.Wrap(err, "ApiServer validation failed")
	}
	if err := c.Defaults.Validate(); err != nil {
		return errors.Wrap(err, "Defaults validation failed")
	}
	if err := c.DNSServer.Validate(); err != nil {
		return errors.Wrap(err, "DNSServer validation failed")
	}
	if err := c.Diagnostics.Validate(); err != nil {
		return errors.Wrap(err, "Diagnostics validation failed")
	}
	if err := c.Experimental.Validate(); err != nil {
		return errors.Wrap(err, "Experimental validation failed")
	}
	if err := c.InterCp.Validate(); err != nil {
		return errors.Wrap(err, "InterCp validation failed")
	}
	if err := c.Tracing.Validate(); err != nil {
		return errors.Wrap(err, "Tracing validation failed")
	}
	if err := c.Policies.Validate(); err != nil {
		return errors.Wrap(err, "Policies validation failed")
	}
	if err := c.IPAM.Validate(); err != nil {
		return errors.Wrap(err, "IPAM validation failed")
	}
	return nil
}

type GeneralConfig struct {
	config.BaseConfig

	// DNSCacheTTL represents duration for how long Kuma CP will cache result of resolving dataplane's domain name
	DNSCacheTTL config_types.Duration `json:"dnsCacheTTL" envconfig:"kuma_general_dns_cache_ttl"`
	// TlsCertFile defines a path to a file with PEM-encoded TLS cert that will be used across all the Kuma Servers.
	TlsCertFile string `json:"tlsCertFile" envconfig:"kuma_general_tls_cert_file"`
	// TlsKeyFile defines a path to a file with PEM-encoded TLS key that will be used across all the Kuma Servers.
	TlsKeyFile string `json:"tlsKeyFile" envconfig:"kuma_general_tls_key_file"`
	// TlsMinVersion defines the minimum TLS version to be used
	TlsMinVersion string `json:"tlsMinVersion" envconfig:"kuma_general_tls_min_version"`
	// TlsMaxVersion defines the maximum TLS version to be used
	TlsMaxVersion string `json:"tlsMaxVersion" envconfig:"kuma_general_tls_max_version"`
	// TlsCipherSuites defines the list of ciphers to use
	TlsCipherSuites []string `json:"tlsCipherSuites" envconfig:"kuma_general_tls_cipher_suites"`
	// WorkDir defines a path to the working directory
	// Kuma stores in this directory autogenerated entities like certificates.
	// If empty then the working directory is $HOME/.kuma
	WorkDir string `json:"workDir" envconfig:"kuma_general_work_dir"`
	// ResilientComponentBaseBackoff configures base backoff for restarting resilient components: KDS sync, Insight resync, PostgresEventListener, etc.
	ResilientComponentBaseBackoff config_types.Duration `json:"resilientComponentBaseBackoff" envconfig:"kuma_general_resilient_component_base_backoff"`
	// ResilientComponentMaxBackoff configures max backoff for restarting resilient component: KDS sync, Insight resync, PostgresEventListener, etc.
	ResilientComponentMaxBackoff config_types.Duration `json:"resilientComponentMaxBackoff" envconfig:"kuma_general_resilient_component_max_backoff"`
}

var _ config.Config = &GeneralConfig{}

func (g *GeneralConfig) Validate() error {
	var errs error
	if g.TlsCertFile == "" && g.TlsKeyFile != "" {
		errs = multierr.Append(errs, errors.New(".TlsCertFile cannot be empty if TlsKeyFile has been set"))
	}
	if g.TlsKeyFile == "" && g.TlsCertFile != "" {
		errs = multierr.Append(errs, errors.New(".TlsKeyFile cannot be empty if TlsCertFile has been set"))
	}
	if _, err := config_types.TLSVersion(g.TlsMinVersion); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsMinVersion"+err.Error()))
	}
	if _, err := config_types.TLSVersion(g.TlsMaxVersion); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsMaxVersion"+err.Error()))
	}
	if _, err := config_types.TLSCiphers(g.TlsCipherSuites); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsCipherSuites"+err.Error()))
	}
	return errs
}

func DefaultGeneralConfig() *GeneralConfig {
	return &GeneralConfig{
		DNSCacheTTL:                   config_types.Duration{Duration: 10 * time.Second},
		WorkDir:                       "",
		TlsCipherSuites:               []string{},
		TlsMinVersion:                 "TLSv1_2",
		ResilientComponentBaseBackoff: config_types.Duration{Duration: 5 * time.Second},
		ResilientComponentMaxBackoff:  config_types.Duration{Duration: 1 * time.Minute},
	}
}

func DefaultDefaultsConfig() *Defaults {
	return &Defaults{
		SkipMeshCreation:           false,
		SkipTenantResources:        false,
		CreateMeshRoutingResources: false,
		SkipHostnameGenerators:     false,
	}
}

type ExperimentalConfig struct {
	config.BaseConfig

	// If true, instead of embedding kubernetes outbounds into Dataplane object, they are persisted next to VIPs in ConfigMap
	// This can improve performance, but it should be enabled only after all instances are migrated to version that supports this config
	KubeOutboundsAsVIPs bool `json:"kubeOutboundsAsVIPs" envconfig:"KUMA_EXPERIMENTAL_KUBE_OUTBOUNDS_AS_VIPS"`
	// Tag first virtual outbound model is compressed version of default Virtual Outbound model
	// It is recommended to use tag first model for deployments with more than 2k services
	// You can enable this flag on existing deployment. In order to downgrade cp with this flag enabled
	// you need to first disable this flag and redeploy cp, after config is rewritten to default
	// format you can downgrade your cp
	UseTagFirstVirtualOutboundModel bool `json:"useTagFirstVirtualOutboundModel" envconfig:"KUMA_EXPERIMENTAL_USE_TAG_FIRST_VIRTUAL_OUTBOUND_MODEL"`
	// List of prefixes that will be used to filter out tags by keys from ingress' available services section.
	// This can trim the size of the ZoneIngress object significantly.
	// The drawback is that you cannot use filtered out tags for traffic routing.
	// If empty, no filter is applied.
	IngressTagFilters []string `json:"ingressTagFilters" envconfig:"KUMA_EXPERIMENTAL_INGRESS_TAG_FILTERS"`
	// KDS event based watchdog settings. It is a more optimal way to generate KDS snapshot config.
	KDSEventBasedWatchdog ExperimentalKDSEventBasedWatchdog `json:"kdsEventBasedWatchdog"`
	// If true then control plane computes reachable services automatically based on MeshTrafficPermission.
	// Lack of MeshTrafficPermission is treated as Deny the traffic.
	AutoReachableServices bool `json:"autoReachableServices" envconfig:"KUMA_EXPERIMENTAL_AUTO_REACHABLE_SERVICES"`
	// Enables sidecar containers in Kubernetes if supported by the Kubernetes
	// environment.
	SidecarContainers bool `json:"sidecarContainers" envconfig:"KUMA_EXPERIMENTAL_SIDECAR_CONTAINERS"`
	// If true then it generates MeshServices from Kubernetes Service.
	GenerateMeshServices bool `json:"generateMeshServices" envconfig:"KUMA_EXPERIMENTAL_GENERATE_MESH_SERVICES"`
	// If true skips persisted VIPs. Change to true only if generateMeshServices is enabled.
	// Do not enable on production.
	SkipPersistedVIPs bool `json:"skipPersistedVIPs" envconfig:"KUMA_EXPERIMENTAL_SKIP_PERSISTED_VIPS"`
	// If true uses Delta xDS to deliver changes to sidecars.
	UseDeltaXds bool `json:"useDeltaXds" envconfig:"KUMA_EXPERIMENTAL_USE_DELTA_XDS"`
}

type ExperimentalKDSEventBasedWatchdog struct {
	// If true, then experimental event based watchdog to generate KDS snapshot is used.
	Enabled bool `json:"enabled" envconfig:"KUMA_EXPERIMENTAL_KDS_EVENT_BASED_WATCHDOG_ENABLED"`
	// How often we flush changes when experimental event based watchdog is used.
	FlushInterval config_types.Duration `json:"flushInterval" envconfig:"KUMA_EXPERIMENTAL_KDS_EVENT_BASED_WATCHDOG_FLUSH_INTERVAL"`
	// How often we schedule full KDS resync when experimental event based watchdog is used.
	FullResyncInterval config_types.Duration `json:"fullResyncInterval" envconfig:"KUMA_EXPERIMENTAL_KDS_EVENT_BASED_WATCHDOG_FULL_RESYNC_INTERVAL"`
	// If true, then initial full resync is going to be delayed by 0 to FullResyncInterval.
	DelayFullResync bool `json:"delayFullResync" envconfig:"KUMA_EXPERIMENTAL_KDS_EVENT_BASED_WATCHDOG_DELAY_FULL_RESYNC"`
}

type IPAMConfig struct {
	MeshService          MeshServiceIPAM          `json:"meshService"`
	MeshExternalService  MeshExternalServiceIPAM  `json:"meshExternalService"`
	MeshMultiZoneService MeshMultiZoneServiceIPAM `json:"meshMultiZoneService"`
	// Interval on which Kuma will allocate new IPs and generate hostnames.
	AllocationInterval config_types.Duration `json:"allocationInterval" envconfig:"KUMA_IPAM_ALLOCATION_INTERVAL"`
}

func (i IPAMConfig) Validate() error {
	if err := i.MeshService.Validate(); err != nil {
		return errors.Wrap(err, "MeshServie validation failed")
	}
	if err := i.MeshExternalService.Validate(); err != nil {
		return errors.Wrap(err, "MeshExternalServie validation failed")
	}
	return nil
}

type MeshServiceIPAM struct {
	// CIDR for MeshService IPs
	CIDR string `json:"cidr" envconfig:"KUMA_IPAM_MESH_SERVICE_CIDR"`
}

func (i MeshServiceIPAM) Validate() error {
	if _, _, err := net.ParseCIDR(i.CIDR); err != nil {
		return errors.Wrap(err, ".MeshServiceCIDR is invalid")
	}
	return nil
}

type MeshExternalServiceIPAM struct {
	// CIDR for MeshExternalService IPs
	CIDR string `json:"cidr" envconfig:"KUMA_IPAM_MESH_EXTERNAL_SERVICE_CIDR"`
}

func (i MeshExternalServiceIPAM) Validate() error {
	if _, _, err := net.ParseCIDR(i.CIDR); err != nil {
		return errors.Wrap(err, ".MeshExternalServiceCIDR is invalid")
	}
	return nil
}

func (c Config) GetEnvoyAdminPort() uint32 {
	if c.BootstrapServer == nil || c.BootstrapServer.Params == nil {
		return 0
	}
	return c.BootstrapServer.Params.AdminPort
}

type MeshMultiZoneServiceIPAM struct {
	// CIDR for MeshMultiZone IPs
	CIDR string `json:"cidr" envconfig:"KUMA_IPAM_MESH_MULTI_ZONE_SERVICE_CIDR"`
}

func (i MeshMultiZoneServiceIPAM) Validate() error {
	if _, _, err := net.ParseCIDR(i.CIDR); err != nil {
		return errors.Wrap(err, ".MeshMultiZoneServiceCIDR is invalid")
	}
	return nil
}

type MeshServiceConfig struct {
	// How often we check whether MeshServices need to be generated from
	// Dataplanes
	GenerationInterval config_types.Duration `json:"generationInterval" envconfig:"KUMA_MESH_SERVICE_GENERATION_INTERVAL"`
	// How long we wait before deleting a MeshService if all Dataplanes are gone
	DeletionGracePeriod config_types.Duration `json:"deletionGracePeriod" envconfig:"KUMA_MESH_SERVICE_DELETION_GRACE_PERIOD"`
}

func (i MeshServiceConfig) Validate() error {
	return nil
}
