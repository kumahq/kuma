package k8s

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	kube_api "k8s.io/apimachinery/pkg/api/resource"

	"github.com/kumahq/kuma/pkg/config"
)

func DefaultKubernetesRuntimeConfig() *KubernetesRuntimeConfig {
	return &KubernetesRuntimeConfig{
		AdmissionServer: AdmissionServerConfig{
			Port: 5443,
		},
		ControlPlaneServiceName: "kuma-control-plane",
		ServiceAccountName:      "system:serviceaccount:kuma-system:kuma-control-plane",
		Injector: Injector{
			CNIEnabled:           false,
			VirtualProbesEnabled: true,
			VirtualProbesPort:    9000,
			SidecarContainer: SidecarContainer{
				RedirectPortInbound:   15006,
				RedirectPortInboundV6: 15010,
				RedirectPortOutbound:  15001,
				DataplaneContainer: DataplaneContainer{
					Image:     "kuma/kuma-dp:latest",
					UID:       5678,
					GID:       5678,
					DrainTime: 30 * time.Second,
					EnvVars:   map[string]string{},

					ReadinessProbe: SidecarReadinessProbe{
						InitialDelaySeconds: 1,
						TimeoutSeconds:      3,
						PeriodSeconds:       5,
						SuccessThreshold:    1,
						FailureThreshold:    12,
					},
					LivenessProbe: SidecarLivenessProbe{
						InitialDelaySeconds: 60,
						TimeoutSeconds:      3,
						PeriodSeconds:       5,
						FailureThreshold:    12,
					},
					Resources: SidecarResources{
						Requests: SidecarResourceRequests{
							CPU:    "50m",
							Memory: "64Mi",
						},
						Limits: SidecarResourceLimits{
							CPU:    "1000m",
							Memory: "512Mi",
						},
					},
				},
			},
			ContainerPatches: []string{},
			InitContainer: InitContainer{
				Image: "kuma/kuma-init:latest",
			},
			SidecarTraffic: SidecarTraffic{
				ExcludeInboundPorts:  []uint32{},
				ExcludeOutboundPorts: []uint32{},
			},
			Exceptions: Exceptions{
				Labels: map[string]string{
					// when using DeploymentConfig instead of Deployment, OpenShift will create an extra deployer Pod for which we don't want to inject Kuma
					"openshift.io/build.name":            "*",
					"openshift.io/deployer-pod-for.name": "*",
				},
			},
			BuiltinDNS: BuiltinDNS{
				Enabled: true,
				Port:    15053,
			},
			EBPF: EBPF{
				Enabled:              false,
				InstanceIPEnvVarName: "INSTANCE_IP",
				BPFFSPath:            "/run/kuma/bpf",
				ProgramsSourcePath:   "/kuma/ebpf",
			},
		},
		MarshalingCacheExpirationTime: 5 * time.Minute,
		NodeTaintController: NodeTaintController{
			Enabled: false,
			CniApp:  "",
		},
	}
}

// Kubernetes-specific configuration
type KubernetesRuntimeConfig struct {
	// Admission WebHook Server implemented by the Control Plane.
	AdmissionServer AdmissionServerConfig `yaml:"admissionServer"`
	// Injector-specific configuration
	Injector Injector `yaml:"injector,omitempty"`
	// MarshalingCacheExpirationTime defines a duration for how long
	// marshaled objects will be stored in the cache. If equal to 0s then
	// cache is turned off
	MarshalingCacheExpirationTime time.Duration `yaml:"marshalingCacheExpirationTime" envconfig:"kuma_runtime_kubernetes_marshaling_cache_expiration_time"`
	// Name of Service Account that is used to run the Control Plane
	ServiceAccountName string `yaml:"serviceAccountName,omitempty" envconfig:"kuma_runtime_kubernetes_service_account_name"`
	// ControlPlaneServiceName defines service name of the Kuma control plane. It is used to point Kuma DP to proper URL.
	ControlPlaneServiceName string `yaml:"controlPlaneServiceName,omitempty" envconfig:"kuma_runtime_kubernetes_control_plane_service_name"`
	// NodeTaintController that prevents applications from scheduling until CNI is ready.
	NodeTaintController NodeTaintController `yaml:"nodeTaintController"`
}

// Configuration of the Admission WebHook Server implemented by the Control Plane.
type AdmissionServerConfig struct {
	// Address the Admission WebHook Server should be listening on.
	Address string `yaml:"address" envconfig:"kuma_runtime_kubernetes_admission_server_address"`
	// Port the Admission WebHook Server should be listening on.
	Port uint32 `yaml:"port" envconfig:"kuma_runtime_kubernetes_admission_server_port"`
	// Directory with a TLS cert and private key for the Admission WebHook Server.
	// TLS certificate file must be named `tls.crt`.
	// TLS key file must be named `tls.key`.
	CertDir string `yaml:"certDir" envconfig:"kuma_runtime_kubernetes_admission_server_cert_dir"`
}

// Injector defines configuration of a Kuma Sidecar Injector.
type Injector struct {
	// SidecarContainer defines configuration of the Kuma sidecar container.
	SidecarContainer SidecarContainer `yaml:"sidecarContainer,omitempty"`
	// InitContainer defines configuration of the Kuma init container.
	InitContainer InitContainer `yaml:"initContainer,omitempty"`
	// ContainerPatches is an optional list of ContainerPatch names which will be applied
	// to init and sidecar containers if workload is not annotated with a patch list.
	ContainerPatches []string `yaml:"containerPatches" envconfig:"kuma_runtime_kubernetes_injector_container_patches"`
	// CNIEnabled if true runs kuma-cp in CNI compatible mode
	CNIEnabled bool `yaml:"cniEnabled" envconfig:"kuma_runtime_kubernetes_injector_cni_enabled"`
	// VirtualProbesEnabled enables automatic converting HttpGet probes to virtual. Virtual probe
	// serves on sub-path of insecure port 'virtualProbesPort',
	// i.e :8080/health/readiness -> :9000/8080/health/readiness where 9000 is virtualProbesPort
	VirtualProbesEnabled bool `yaml:"virtualProbesEnabled" envconfig:"kuma_runtime_kubernetes_virtual_probes_enabled"`
	// VirtualProbesPort is a port for exposing virtual probes which are not secured by mTLS
	VirtualProbesPort uint32 `yaml:"virtualProbesPort" envconfig:"kuma_runtime_kubernetes_virtual_probes_port"`
	// SidecarTraffic is a configuration for a traffic that is intercepted by sidecar
	SidecarTraffic SidecarTraffic `yaml:"sidecarTraffic"`
	// Exceptions defines list of exceptions for Kuma injection
	Exceptions Exceptions `yaml:"exceptions"`
	// CaCertFile is CA certificate which will be used to verify a connection to the control plane
	CaCertFile string     `yaml:"caCertFile" envconfig:"kuma_runtime_kubernetes_injector_ca_cert_file"`
	BuiltinDNS BuiltinDNS `yaml:"builtinDNS"`
	// EBPF is a configuration for ebpf if transparent proxy should be installed
	// using ebpf instead of iptables
	EBPF EBPF `yaml:"ebpf"`
}

// Exceptions defines list of exceptions for Kuma injection
type Exceptions struct {
	// Labels is a map of labels for exception. If pod matches label with given value Kuma won't be injected. Specify '*' to match any value.
	Labels map[string]string `yaml:"labels" envconfig:"kuma_runtime_kubernetes_exceptions_labels"`
}

type SidecarTraffic struct {
	// List of inbound ports that will be excluded from interception.
	// This setting is applied on every pod unless traffic.kuma.io/exclude-inbound-ports annotation is specified on Pod.
	ExcludeInboundPorts []uint32 `yaml:"excludeInboundPorts" envconfig:"kuma_runtime_kubernetes_sidecar_traffic_exclude_inbound_ports"`
	// List of outbound ports that will be excluded from interception.
	// This setting is applied on every pod unless traffic.kuma.io/exclude-oubound-ports annotation is specified on Pod.
	ExcludeOutboundPorts []uint32 `yaml:"excludeOutboundPorts" envconfig:"kuma_runtime_kubernetes_sidecar_traffic_exclude_outbound_ports"`
}

// DataplaneContainer defines the configuration of a Kuma dataplane proxy container.
type DataplaneContainer struct {
	// Image name.
	Image string `yaml:"image,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_image"`
	// User ID.
	UID int64 `yaml:"uid,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_uid"`
	// Group ID.
	GID int64 `yaml:"gid,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_gui"`
	// Deprecated: Use KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT instead.
	AdminPort uint32 `yaml:"adminPort,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_admin_port"`
	// Drain time for listeners.
	DrainTime time.Duration `yaml:"drainTime,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_drain_time"`
	// Readiness probe.
	ReadinessProbe SidecarReadinessProbe `yaml:"readinessProbe,omitempty"`
	// Liveness probe.
	LivenessProbe SidecarLivenessProbe `yaml:"livenessProbe,omitempty"`
	// Compute resource requirements.
	Resources SidecarResources `yaml:"resources,omitempty"`
	// EnvVars are additional environment variables that can be placed on Kuma DP sidecar
	EnvVars map[string]string `yaml:"envVars" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_env_vars"`
}

// SidecarContainer defines configuration of the Kuma sidecar container.
type SidecarContainer struct {
	DataplaneContainer `yaml:",inline"`
	// Redirect port for inbound traffic.
	RedirectPortInbound uint32 `yaml:"redirectPortInbound,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_redirect_port_inbound"`
	// Redirect port for inbound IPv6 traffic.
	RedirectPortInboundV6 uint32 `yaml:"redirectPortInboundV6,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_redirect_port_inbound_v6"`
	// Redirect port for outbound traffic.
	RedirectPortOutbound uint32 `yaml:"redirectPortOutbound,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_redirect_port_outbound"`
}

// SidecarReadinessProbe defines periodic probe of container service readiness.
type SidecarReadinessProbe struct {
	// Number of seconds after the container has started before readiness probes are initiated.
	InitialDelaySeconds int32 `yaml:"initialDelaySeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_readiness_probe_initial_delay_seconds"`
	// Number of seconds after which the probe times out.
	TimeoutSeconds int32 `yaml:"timeoutSeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_readiness_probe_timeout_seconds"`
	// Number of seconds after which the probe times out.
	PeriodSeconds int32 `yaml:"periodSeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_readiness_probe_period_seconds"`
	// Minimum consecutive successes for the probe to be considered successful after having failed.
	SuccessThreshold int32 `yaml:"successThreshold,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_readiness_probe_success_threshold"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	FailureThreshold int32 `yaml:"failureThreshold,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_readiness_probe_failure_threshold"`
}

// SidecarLivenessProbe defines periodic probe of container service liveness.
type SidecarLivenessProbe struct {
	// Number of seconds after the container has started before liveness probes are initiated.
	InitialDelaySeconds int32 `yaml:"initialDelaySeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_liveness_probe_initial_delay_seconds"`
	// Number of seconds after which the probe times out.
	TimeoutSeconds int32 `yaml:"timeoutSeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_liveness_probe_timeout_seconds"`
	// How often (in seconds) to perform the probe.
	PeriodSeconds int32 `yaml:"periodSeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_liveness_probe_period_seconds"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	FailureThreshold int32 `yaml:"failureThreshold,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_liveness_probe_failure_threshold"`
}

// SidecarResources defines compute resource requirements.
type SidecarResources struct {
	// Minimum amount of compute resources required.
	Requests SidecarResourceRequests `yaml:"requests,omitempty"`
	// Maximum amount of compute resources allowed.
	Limits SidecarResourceLimits `yaml:"limits,omitempty"`
}

// SidecarResourceRequests defines the minimum amount of compute resources required.
type SidecarResourceRequests struct {
	// CPU, in cores. (500m = .5 cores)
	CPU string `yaml:"cpu,omitempty" envconfig:"kuma_injector_sidecar_container_resources_requests_cpu"`
	// Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	Memory string `yaml:"memory,omitempty" envconfig:"kuma_injector_sidecar_container_resources_requests_memory"`
}

// SidecarResourceLimits defines the maximum amount of compute resources allowed.
type SidecarResourceLimits struct {
	// CPU, in cores. (500m = .5 cores)
	CPU string `yaml:"cpu,omitempty" envconfig:"kuma_injector_sidecar_container_resources_limits_cpu"`
	// Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	Memory string `yaml:"memory,omitempty" envconfig:"kuma_injector_sidecar_container_resources_limits_memory"`
}

// InitContainer defines configuration of the Kuma init container.
type InitContainer struct {
	// Image name.
	Image string `yaml:"image,omitempty" envconfig:"kuma_injector_init_container_image"`
}

type BuiltinDNS struct {
	// Use the built-in DNS
	Enabled bool `yaml:"enabled,omitempty" envconfig:"kuma_runtime_kubernetes_injector_builtin_dns_enabled"`
	// Redirect port for DNS
	Port uint32 `yaml:"port,omitempty" envconfig:"kuma_runtime_kubernetes_injector_builtin_dns_port"`
}

// EBPF defines configuration for the ebpf, when transparent proxy is marked to be
// installed using ebpf instead of iptables
type EBPF struct {
	// Install transparent proxy using ebpf
	Enabled bool `yaml:"enabled" envconfig:"kuma_runtime_kubernetes_injector_ebpf_enabled"`
	// Name of the environmental variable which will include IP address of the pod
	InstanceIPEnvVarName string `yaml:"instanceIPEnvVarName,omitempty" envconfig:"kuma_runtime_kubernetes_injector_ebpf_instance_ip_env_var_name"`
	// Path where BPF file system will be mounted for pinning ebpf programs and maps
	BPFFSPath string `yaml:"bpffsPath,omitempty" envconfig:"kuma_runtime_kubernetes_injector_ebpf_bpffs_path"`
	// Path where compiled eBPF programs are placed
	ProgramsSourcePath string `yaml:"programsSourcePath,omitempty" envconfig:"kuma_runtime_kubernetes_injector_ebpf_programs_source_path"`
}

type NodeTaintController struct {
	// If true enables the taint controller.
	Enabled bool `yaml:"enabled" envconfig:"kuma_runtime_kubernetes_node_taint_controller_enabled"`
	// Value of app label on CNI pod that indicates if node can be ready.
	CniApp string `yaml:"cniApp" envconfig:"kuma_runtime_kubernetes_node_taint_controller_cni_app"`
}

func (n *NodeTaintController) Validate() error {
	if n.Enabled && n.CniApp == "" {
		return errors.New(".CniApp has to be set when .Enabled is true")
	}
	return nil
}

var _ config.Config = &KubernetesRuntimeConfig{}

func (c *KubernetesRuntimeConfig) Sanitize() {
}

func (c *KubernetesRuntimeConfig) Validate() (errs error) {
	if err := c.AdmissionServer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".AdmissionServer is not valid"))
	}
	if err := c.Injector.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Injector is not valid"))
	}
	if c.MarshalingCacheExpirationTime < 0 {
		errs = multierr.Append(errs, errors.Errorf(".MarshalingCacheExpirationTime must be positive or equal to 0"))
	}
	return
}

var _ config.Config = &AdmissionServerConfig{}

func (c *AdmissionServerConfig) Sanitize() {
}

func (c *AdmissionServerConfig) Validate() (errs error) {
	if 65535 < c.Port {
		errs = multierr.Append(errs, errors.Errorf(".Port must be in the range [0, 65535]"))
	}
	if c.CertDir == "" {
		errs = multierr.Append(errs, errors.Errorf(".CertDir should not be empty"))
	}
	return
}

var _ config.Config = &Injector{}

func (i *Injector) Sanitize() {
	i.InitContainer.Sanitize()
	i.SidecarContainer.Sanitize()
}

func (i *Injector) Validate() (errs error) {
	if err := i.SidecarContainer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".SidecarContainer is not valid"))
	}
	if err := i.InitContainer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".InitContainer is not valid"))
	}
	return
}

var _ config.Config = &SidecarContainer{}

func (c *SidecarContainer) Sanitize() {
	c.Resources.Sanitize()
	c.LivenessProbe.Sanitize()
	c.ReadinessProbe.Sanitize()
}

func (c *SidecarContainer) Validate() (errs error) {
	if c.Image == "" {
		errs = multierr.Append(errs, errors.Errorf(".Image must be non-empty"))
	}
	if 65535 < c.RedirectPortInbound {
		errs = multierr.Append(errs, errors.Errorf(".RedirectPortInbound must be in the range [0, 65535]"))
	}
	if 0 != c.RedirectPortInboundV6 && 65535 < c.RedirectPortInboundV6 {
		errs = multierr.Append(errs, errors.Errorf(".RedirectPortInboundV6 must be in the range [0, 65535]"))
	}
	if 65535 < c.RedirectPortOutbound {
		errs = multierr.Append(errs, errors.Errorf(".RedirectPortOutbound must be in the range [0, 65535]"))
	}
	if 65535 < c.AdminPort {
		errs = multierr.Append(errs, errors.Errorf(".AdminPort must be in the range [0, 65535]"))
	}
	if c.DrainTime <= 0 {
		errs = multierr.Append(errs, errors.Errorf(".DrainTime must be positive"))
	}
	if err := c.ReadinessProbe.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".ReadinessProbe is not valid"))
	}
	if err := c.LivenessProbe.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".LivenessProbe is not valid"))
	}
	if err := c.Resources.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Resources is not valid"))
	}
	return
}

var _ config.Config = &InitContainer{}

func (c *InitContainer) Sanitize() {
}

func (c *InitContainer) Validate() (errs error) {
	if c.Image == "" {
		errs = multierr.Append(errs, errors.Errorf(".Image must be non-empty"))
	}
	return
}

var _ config.Config = &SidecarReadinessProbe{}

func (c *SidecarReadinessProbe) Sanitize() {
}

func (c *SidecarReadinessProbe) Validate() (errs error) {
	if c.InitialDelaySeconds < 1 {
		errs = multierr.Append(errs, errors.Errorf(".InitialDelaySeconds must be >= 1"))
	}
	if c.TimeoutSeconds < 1 {
		errs = multierr.Append(errs, errors.Errorf(".TimeoutSeconds must be >= 1"))
	}
	if c.PeriodSeconds < 1 {
		errs = multierr.Append(errs, errors.Errorf(".PeriodSeconds must be >= 1"))
	}
	if c.SuccessThreshold < 1 {
		errs = multierr.Append(errs, errors.Errorf(".SuccessThreshold must be >= 1"))
	}
	if c.FailureThreshold < 1 {
		errs = multierr.Append(errs, errors.Errorf(".FailureThreshold must be >= 1"))
	}
	return
}

var _ config.Config = &SidecarLivenessProbe{}

func (c *SidecarLivenessProbe) Sanitize() {
}

func (c *SidecarLivenessProbe) Validate() (errs error) {
	if c.InitialDelaySeconds < 1 {
		errs = multierr.Append(errs, errors.Errorf(".InitialDelaySeconds must be >= 1"))
	}
	if c.TimeoutSeconds < 1 {
		errs = multierr.Append(errs, errors.Errorf(".TimeoutSeconds must be >= 1"))
	}
	if c.PeriodSeconds < 1 {
		errs = multierr.Append(errs, errors.Errorf(".PeriodSeconds must be >= 1"))
	}
	if c.FailureThreshold < 1 {
		errs = multierr.Append(errs, errors.Errorf(".FailureThreshold must be >= 1"))
	}
	return
}

var _ config.Config = &SidecarResources{}

func (c *SidecarResources) Sanitize() {
	c.Limits.Sanitize()
	c.Requests.Sanitize()
}

func (c *SidecarResourceRequests) Sanitize() {
}

func (c *SidecarResources) Validate() (errs error) {
	if err := c.Requests.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Requests is not valid"))
	}
	if err := c.Limits.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Limits is not valid"))
	}
	return
}

var _ config.Config = &SidecarResourceRequests{}

func (c *SidecarResourceRequests) Validate() (errs error) {
	if _, err := kube_api.ParseQuantity(c.CPU); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".CPU is not valid"))
	}
	if _, err := kube_api.ParseQuantity(c.Memory); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Memory is not valid"))
	}
	return
}

var _ config.Config = &SidecarResourceLimits{}

func (c *SidecarResourceLimits) Sanitize() {
}

func (c *SidecarResourceLimits) Validate() (errs error) {
	if _, err := kube_api.ParseQuantity(c.CPU); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".CPU is not valid"))
	}
	if _, err := kube_api.ParseQuantity(c.Memory); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Memory is not valid"))
	}
	return
}

var _ config.Config = &BuiltinDNS{}

func (c *BuiltinDNS) Sanitize() {
}

func (c *BuiltinDNS) Validate() (errs error) {
	if 65535 < c.Port {
		errs = multierr.Append(errs, errors.Errorf(".port must be in the range [0, 65535]"))
	}
	return
}
