package k8s

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	kube_api "k8s.io/apimachinery/pkg/api/resource"

	"github.com/kumahq/kuma/pkg/config"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/core"
)

const defaultServiceAccountName = "system:serviceaccount:kuma-system:kuma-control-plane"

var logger = core.Log.WithName("kubernetes-config")

func DefaultKubernetesRuntimeConfig() *KubernetesRuntimeConfig {
	return &KubernetesRuntimeConfig{
		AdmissionServer: AdmissionServerConfig{
			Port: 5443,
		},
		ControlPlaneServiceName: "kuma-control-plane",
		ServiceAccountName:      defaultServiceAccountName,
		Injector: Injector{
			CNIEnabled:                false,
			VirtualProbesEnabled:      true,
			VirtualProbesPort:         9000,
			ApplicationProbeProxyPort: 9000,
			SidecarContainer: SidecarContainer{
				IpFamilyMode:         "dualstack",
				RedirectPortInbound:  15006,
				RedirectPortOutbound: 15001,
				DataplaneContainer: DataplaneContainer{
					Image:     "kuma/kuma-dp:latest",
					UID:       5678,
					GID:       5678,
					DrainTime: config_types.Duration{Duration: 30 * time.Second},
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
					StartupProbe: SidecarStartupProbe{
						InitialDelaySeconds: 1,
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
				ExcludeInboundIPs:    []string{},
				ExcludeOutboundIPs:   []string{},
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
				Logging: false,
			},
			EBPF: EBPF{
				Enabled:              false,
				InstanceIPEnvVarName: "INSTANCE_IP",
				BPFFSPath:            "/sys/fs/bpf",
				CgroupPath:           "/sys/fs/cgroup",
				ProgramsSourcePath:   "/tmp/kuma-ebpf",
			},
			IgnoredServiceSelectorLabels: []string{},
			// topology labels that are useful for, for example, MeshLoadBalancingStrategy policy.
			NodeLabelsToCopy: []string{"topology.kubernetes.io/zone", "topology.kubernetes.io/region", "kubernetes.io/hostname"},
		},
		MarshalingCacheExpirationTime: config_types.Duration{Duration: 5 * time.Minute},
		NodeTaintController: NodeTaintController{
			Enabled:      false,
			CniApp:       "",
			CniNamespace: "kube-system",
		},
		ControllersConcurrency: ControllersConcurrency{
			PodController: 10,
		},
		ClientConfig: ClientConfig{
			Qps:      100,
			BurstQps: 100,
		},
		LeaderElection: LeaderElection{
			LeaseDuration: config_types.Duration{Duration: 15 * time.Second},
			RenewDeadline: config_types.Duration{Duration: 10 * time.Second},
		},
	}
}

// KubernetesRuntimeConfig defines Kubernetes-specific configuration
type KubernetesRuntimeConfig struct {
	config.BaseConfig

	// Admission WebHook Server implemented by the Control Plane.
	AdmissionServer AdmissionServerConfig `json:"admissionServer"`
	// Injector-specific configuration
	Injector Injector `json:"injector,omitempty"`
	// MarshalingCacheExpirationTime defines a duration for how long
	// marshaled objects will be stored in the cache. If equal to 0s then
	// cache is turned off
	MarshalingCacheExpirationTime config_types.Duration `json:"marshalingCacheExpirationTime" envconfig:"kuma_runtime_kubernetes_marshaling_cache_expiration_time"`
	// Name of Service Account that is used to run the Control Plane
	// Deprecated: Use AllowedUsers instead.
	ServiceAccountName string `json:"serviceAccountName,omitempty" envconfig:"kuma_runtime_kubernetes_service_account_name"`
	// List of names of Service Accounts that admission requests are allowed.
	// This list is appended with Control Plane's Service Account and generic-garbage-collector
	AllowedUsers []string `json:"allowedUsers,omitempty" envconfig:"kuma_runtime_kubernetes_allowed_users"`
	// ControlPlaneServiceName defines service name of the Kuma control plane. It is used to point Kuma DP to proper URL.
	ControlPlaneServiceName string `json:"controlPlaneServiceName,omitempty" envconfig:"kuma_runtime_kubernetes_control_plane_service_name"`
	// NodeTaintController that prevents applications from scheduling until CNI is ready.
	NodeTaintController NodeTaintController `json:"nodeTaintController"`
	// Kubernetes' resources reconciliation concurrency configuration
	ControllersConcurrency ControllersConcurrency `json:"controllersConcurrency"`
	// Kubernetes client configuration
	ClientConfig ClientConfig `json:"clientConfig"`
	// Kubernetes leader election configuration
	LeaderElection LeaderElection `json:"leaderElection"`
	// SkipMeshOwnerReference is a flag that allows to skip adding Mesh owner reference to resources.
	// If this is set to true, deleting a Mesh will not delete resources that belong to that Mesh.
	// This can be useful when resources are managed in Argo CD where creation/deletion is managed there.
	SkipMeshOwnerReference bool `json:"skipMeshOwnerReference" envconfig:"kuma_runtime_kubernetes_skip_mesh_owner_reference"`
	// If true, then control plane can support TLS secrets for builtin gateway outside of mesh system namespace.
	// The downside is that control plane requires permission to read Secrets in all namespaces.
	SupportGatewaySecretsInAllNamespaces bool `json:"supportGatewaySecretsInAllNamespaces" envconfig:"kuma_runtime_kubernetes_support_gateway_secrets_in_all_namespaces"`
}

type ControllersConcurrency struct {
	// PodController defines maximum concurrent reconciliations of Pod resources
	// Default value 10. If set to 0 kube controller-runtime default value of 1 will be used.
	PodController int `json:"podController" envconfig:"kuma_runtime_kubernetes_controllers_concurrency_pod_controller"`
}

type ClientConfig struct {
	// Qps defines maximum requests kubernetes client is allowed to make per second.
	// Default value 100. If set to 0 kube-client default value of 5 will be used.
	Qps int `json:"qps" envconfig:"kuma_runtime_kubernetes_client_config_qps"`
	// BurstQps defines maximum burst requests kubernetes client is allowed to make per second
	// Default value 100. If set to 0 kube-client default value of 10 will be used.
	BurstQps int `json:"burstQps" envconfig:"kuma_runtime_kubernetes_client_config_burst_qps"`
}

type LeaderElection struct {
	// LeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership. This is measured against time of
	// last observed ack. Default is 15 seconds.
	LeaseDuration config_types.Duration `json:"leaseDuration" envconfig:"kuma_runtime_kubernetes_leader_election_lease_duration"`
	// RenewDeadline is the duration that the acting controlplane will retry
	// refreshing leadership before giving up. Default is 10 seconds.
	RenewDeadline config_types.Duration `json:"renewDeadline" envconfig:"kuma_runtime_kubernetes_leader_election_renew_deadline"`
}

// AdmissionServerConfig defines configuration of the Admission WebHook Server implemented by
// the Control Plane.
type AdmissionServerConfig struct {
	config.BaseConfig

	// Address the Admission WebHook Server should be listening on.
	Address string `json:"address" envconfig:"kuma_runtime_kubernetes_admission_server_address"`
	// Port the Admission WebHook Server should be listening on.
	Port uint32 `json:"port" envconfig:"kuma_runtime_kubernetes_admission_server_port"`
	// Directory with a TLS cert and private key for the Admission WebHook Server.
	// TLS certificate file must be named `tls.crt`.
	// TLS key file must be named `tls.key`.
	CertDir string `json:"certDir" envconfig:"kuma_runtime_kubernetes_admission_server_cert_dir"`
}

// Injector defines configuration of a Kuma Sidecar Injector.
type Injector struct {
	// SidecarContainer defines configuration of the Kuma sidecar container.
	SidecarContainer SidecarContainer `json:"sidecarContainer,omitempty"`
	// InitContainer defines configuration of the Kuma init container.
	InitContainer InitContainer `json:"initContainer,omitempty"`
	// ContainerPatches is an optional list of ContainerPatch names which will be applied
	// to init and sidecar containers if workload is not annotated with a patch list.
	ContainerPatches []string `json:"containerPatches" envconfig:"kuma_runtime_kubernetes_injector_container_patches"`
	// CNIEnabled if true runs kuma-cp in CNI compatible mode
	CNIEnabled bool `json:"cniEnabled" envconfig:"kuma_runtime_kubernetes_injector_cni_enabled"`
	// VirtualProbesEnabled enables automatic converting pod probes to virtual probes that is proxied by the sidecar.
	VirtualProbesEnabled bool `json:"virtualProbesEnabled" envconfig:"kuma_runtime_kubernetes_virtual_probes_enabled"`
	// VirtualProbesPort is a port for exposing virtual probes which are not secured by mTLS.
	VirtualProbesPort uint32 `json:"virtualProbesPort" envconfig:"kuma_runtime_kubernetes_virtual_probes_port"`
	// ApplicationProbeProxyPort is a port for proxying application probes, it is not secured by mTLS.
	ApplicationProbeProxyPort uint32 `json:"applicationProbeProxyPort" envconfig:"kuma_runtime_kubernetes_application_probe_proxy_port"`
	// SidecarTraffic is a configuration for traffic that is intercepted by sidecar
	SidecarTraffic SidecarTraffic `json:"sidecarTraffic"`
	// Exceptions defines list of exceptions for Kuma injection
	Exceptions Exceptions `json:"exceptions"`
	// CaCertFile is CA certificate which will be used to verify a connection to the control plane
	CaCertFile string     `json:"caCertFile" envconfig:"kuma_runtime_kubernetes_injector_ca_cert_file"`
	BuiltinDNS BuiltinDNS `json:"builtinDNS"`
	// EBPF is a configuration for ebpf if transparent proxy should be installed
	// using ebpf instead of iptables
	EBPF EBPF `json:"ebpf"`
	// IgnoredServiceSelectorLabels defines a list ignored labels in Service selector.
	// If Pod matches a Service with ignored labels, but does not match it fully, it gets Ignored inbound.
	// It is useful when you change Service selector and expect traffic to be sent immediately.
	// An example of this is ArgoCD's BlueGreen deployment and "rollouts-pod-template-hash" selector.
	IgnoredServiceSelectorLabels []string `json:"ignoredServiceSelectorLabels" envconfig:"KUMA_RUNTIME_KUBERNETES_INJECTOR_IGNORED_SERVICE_SELECTOR_LABELS"`
	// NodeLabelsToCopy defines a list of node labels that should be copied to the Pod.
	NodeLabelsToCopy []string `json:"nodeLabelsToCopy" envconfig:"KUMA_RUNTIME_KUBERNETES_INJECTOR_NODE_LABELS_TO_COPY"`
}

// Exceptions defines list of exceptions for Kuma injection
type Exceptions struct {
	// Labels is a map of labels for exception. If pod matches label with given value Kuma won't be injected. Specify '*' to match any value.
	Labels map[string]string `json:"labels" envconfig:"kuma_runtime_kubernetes_exceptions_labels"`
}

type SidecarTraffic struct {
	// List of inbound ports that will be excluded from interception.
	// This setting is applied on every pod unless traffic.kuma.io/exclude-inbound-ports annotation is specified on Pod.
	ExcludeInboundPorts []uint32 `json:"excludeInboundPorts" envconfig:"kuma_runtime_kubernetes_sidecar_traffic_exclude_inbound_ports"`
	// List of outbound ports that will be excluded from interception.
	// This setting is applied on every pod unless traffic.kuma.io/exclude-outbound-ports annotation is specified on Pod.
	ExcludeOutboundPorts []uint32 `json:"excludeOutboundPorts" envconfig:"kuma_runtime_kubernetes_sidecar_traffic_exclude_outbound_ports"`
	// List of inbound IP addresses that will be excluded from interception.
	// This setting is applied on every pod unless traffic.kuma.io/exclude-inbound-ips annotation is specified on the Pod.
	// IP addresses can be specified with or without CIDR notation, and multiple addresses can be separated by commas.
	ExcludeInboundIPs []string `json:"excludeInboundIPs" envconfig:"kuma_runtime_kubernetes_sidecar_traffic_exclude_inbound_ips"`
	// List of outbound IP addresses that will be excluded from interception.
	// This setting is applied on every pod unless traffic.kuma.io/exclude-outbound-ips annotation is specified on the Pod.
	// IP addresses can be specified with or without CIDR notation, and multiple addresses can be separated by commas.
	ExcludeOutboundIPs []string `json:"excludeOutboundIPs" envconfig:"kuma_runtime_kubernetes_sidecar_traffic_exclude_outbound_ips"`
}

// DataplaneContainer defines the configuration of a Kuma dataplane proxy container.
type DataplaneContainer struct {
	// Image name.
	Image string `json:"image,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_image"`
	// User ID.
	UID int64 `json:"uid,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_uid"`
	// Group ID.
	GID int64 `json:"gid,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_gui"`
	// Deprecated: Use KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT instead.
	AdminPort uint32 `json:"adminPort,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_admin_port"`
	// Drain time for listeners.
	DrainTime config_types.Duration `json:"drainTime,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_drain_time"`
	// Readiness probe.
	ReadinessProbe SidecarReadinessProbe `json:"readinessProbe,omitempty"`
	// Liveness probe.
	LivenessProbe SidecarLivenessProbe `json:"livenessProbe,omitempty"`
	// Startup probe for sidecar containers feature
	StartupProbe SidecarStartupProbe `json:"startupProbe,omitempty"`
	// Compute resource requirements.
	Resources SidecarResources `json:"resources,omitempty"`
	// EnvVars are additional environment variables that can be placed on Kuma DP sidecar
	EnvVars map[string]string `json:"envVars" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_env_vars"`
}

// SidecarContainer defines configuration of the Kuma sidecar container.
type SidecarContainer struct {
	DataplaneContainer `json:",inline"`
	// Redirect port for inbound traffic.
	RedirectPortInbound uint32 `json:"redirectPortInbound,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_redirect_port_inbound"`
	// The IP family mode to enable traffic redirection for. Can be "ipv4" or "dualstack".
	IpFamilyMode string `json:"ipFamilyMode,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_ip_family_mode"`
	// Redirect port for outbound traffic.
	RedirectPortOutbound uint32 `json:"redirectPortOutbound,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_redirect_port_outbound"`
	// WaitForDataplaneReady enables a script that waits until Envoy is ready.
	// With the current Kubernetes behavior, any other container in the Pod will wait until the script is complete.
	WaitForDataplaneReady bool `json:"waitForDataplaneReady" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_wait_for_dataplane_ready"`
}

// SidecarReadinessProbe defines periodic probe of container service readiness.
type SidecarReadinessProbe struct {
	config.BaseConfig

	// Number of seconds after the container has started before readiness probes are initiated.
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_readiness_probe_initial_delay_seconds"`
	// Number of seconds after which the probe times out.
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_readiness_probe_timeout_seconds"`
	// Number of seconds after which the probe times out.
	PeriodSeconds int32 `json:"periodSeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_readiness_probe_period_seconds"`
	// Minimum consecutive successes for the probe to be considered successful after having failed.
	SuccessThreshold int32 `json:"successThreshold,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_readiness_probe_success_threshold"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	FailureThreshold int32 `json:"failureThreshold,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_readiness_probe_failure_threshold"`
}

// SidecarLivenessProbe defines periodic probe of container service liveness.
type SidecarLivenessProbe struct {
	config.BaseConfig

	// Number of seconds after the container has started before liveness probes are initiated.
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_liveness_probe_initial_delay_seconds"`
	// Number of seconds after which the probe times out.
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_liveness_probe_timeout_seconds"`
	// How often (in seconds) to perform the probe.
	PeriodSeconds int32 `json:"periodSeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_liveness_probe_period_seconds"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	FailureThreshold int32 `json:"failureThreshold,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_liveness_probe_failure_threshold"`
}

// SidecarStartupProbe defines startup probe of Kuma sidecar.
type SidecarStartupProbe struct {
	config.BaseConfig

	// Number of seconds after the container has started before startup probes are initiated.
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_startup_probe_initial_delay_seconds"`
	// Number of seconds after which the probe times out.
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_startup_probe_timeout_seconds"`
	// How often (in seconds) to perform the probe.
	PeriodSeconds int32 `json:"periodSeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_startup_probe_period_seconds"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	FailureThreshold int32 `json:"failureThreshold,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_startup_probe_failure_threshold"`
}

// SidecarResources defines compute resource requirements.
type SidecarResources struct {
	// Minimum amount of compute resources required.
	Requests SidecarResourceRequests `json:"requests,omitempty"`
	// Maximum amount of compute resources allowed.
	Limits SidecarResourceLimits `json:"limits,omitempty"`
}

// SidecarResourceRequests defines the minimum amount of compute resources required.
type SidecarResourceRequests struct {
	config.BaseConfig

	// CPU, in cores. (500m = .5 cores)
	CPU string `json:"cpu,omitempty" envconfig:"kuma_injector_sidecar_container_resources_requests_cpu"`
	// Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	Memory string `json:"memory,omitempty" envconfig:"kuma_injector_sidecar_container_resources_requests_memory"`
}

// SidecarResourceLimits defines the maximum amount of compute resources allowed.
type SidecarResourceLimits struct {
	config.BaseConfig

	// CPU, in cores. (500m = .5 cores)
	CPU string `json:"cpu,omitempty" envconfig:"kuma_injector_sidecar_container_resources_limits_cpu"`
	// Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	Memory string `json:"memory,omitempty" envconfig:"kuma_injector_sidecar_container_resources_limits_memory"`
}

// InitContainer defines configuration of the Kuma init container.
type InitContainer struct {
	config.BaseConfig

	// Image name.
	Image string `json:"image,omitempty" envconfig:"kuma_injector_init_container_image"`
}

type BuiltinDNS struct {
	config.BaseConfig

	// Use the built-in DNS
	Enabled bool `json:"enabled,omitempty" envconfig:"kuma_runtime_kubernetes_injector_builtin_dns_enabled"`
	// Redirect port for DNS
	Port uint32 `json:"port,omitempty" envconfig:"kuma_runtime_kubernetes_injector_builtin_dns_port"`
	// Turn on query logging for DNS
	Logging bool `json:"logging,omitempty" envconfig:"kuma_runtime_kubernetes_injector_builtin_dns_logging"`
}

// EBPF defines configuration for the ebpf, when transparent proxy is marked to be
// installed using ebpf instead of iptables
type EBPF struct {
	// Install transparent proxy using ebpf
	Enabled bool `json:"enabled" envconfig:"kuma_runtime_kubernetes_injector_ebpf_enabled"`
	// Name of the environmental variable which will include IP address of the pod
	InstanceIPEnvVarName string `json:"instanceIPEnvVarName,omitempty" envconfig:"kuma_runtime_kubernetes_injector_ebpf_instance_ip_env_var_name"`
	// Path where BPF file system will be mounted for pinning ebpf programs and maps
	BPFFSPath string `json:"bpffsPath,omitempty" envconfig:"kuma_runtime_kubernetes_injector_ebpf_bpffs_path"`
	// Path of mounted cgroup2
	CgroupPath string `json:"cgroupPath,omitempty" envconfig:"kuma_runtime_kubernetes_injector_ebpf_cgroup_path"`
	// Name of the network interface which should be used to attach to it TC programs
	// when not specified, we will try to automatically determine it
	TCAttachIface string `json:"tcAttachIface,omitempty" envconfig:"kuma_runtime_kubernetes_injector_ebpf_tc_attach_iface"`
	// Path where compiled eBPF programs are placed
	ProgramsSourcePath string `json:"programsSourcePath,omitempty" envconfig:"kuma_runtime_kubernetes_injector_ebpf_programs_source_path"`
}

type NodeTaintController struct {
	// If true enables the taint controller.
	Enabled bool `json:"enabled" envconfig:"kuma_runtime_kubernetes_node_taint_controller_enabled"`
	// Value of app label on CNI pod that indicates if node can be ready.
	CniApp string `json:"cniApp" envconfig:"kuma_runtime_kubernetes_node_taint_controller_cni_app"`
	// Value of CNI namespace.
	CniNamespace string `json:"cniNamespace" envconfig:"kuma_runtime_kubernetes_node_taint_controller_cni_namespace"`
}

func (n *NodeTaintController) Validate() error {
	if n.Enabled && n.CniApp == "" {
		return errors.New(".CniApp has to be set when .Enabled is true")
	}
	return nil
}

var _ config.Config = &KubernetesRuntimeConfig{}

func (c *KubernetesRuntimeConfig) PostProcess() error {
	return multierr.Combine(
		c.AdmissionServer.PostProcess(),
		c.Injector.PostProcess(),
	)
}

func (c *KubernetesRuntimeConfig) Validate() error {
	var errs error
	if err := c.AdmissionServer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".AdmissionServer is not valid"))
	}
	if err := c.Injector.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Injector is not valid"))
	}
	if c.MarshalingCacheExpirationTime.Duration < 0 {
		errs = multierr.Append(errs, errors.Errorf(".MarshalingCacheExpirationTime must be positive or equal to 0"))
	}
	if c.ServiceAccountName != defaultServiceAccountName {
		logger.Info("[WARNING]: using deprecated configuration option - .ServiceAccountName, please use AllowedUsers.")
	}
	return errs
}

var _ config.Config = &AdmissionServerConfig{}

func (c *AdmissionServerConfig) Validate() error {
	var errs error
	if 65535 < c.Port {
		errs = multierr.Append(errs, errors.Errorf(".Port must be in the range [0, 65535]"))
	}
	if c.CertDir == "" {
		errs = multierr.Append(errs, errors.Errorf(".CertDir should not be empty"))
	}
	return errs
}

var _ config.Config = &Injector{}

func (i *Injector) Sanitize() {
	i.InitContainer.Sanitize()
	i.SidecarContainer.Sanitize()
}

func (i *Injector) PostProcess() error {
	return multierr.Combine(
		i.InitContainer.PostProcess(),
		i.SidecarContainer.PostProcess(),
	)
}

func (i *Injector) Validate() error {
	var errs error
	if err := i.SidecarContainer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".SidecarContainer is not valid"))
	}
	if err := i.InitContainer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".InitContainer is not valid"))
	}
	return errs
}

var _ config.Config = &SidecarContainer{}

func (c *SidecarContainer) Sanitize() {
	c.Resources.Sanitize()
	c.LivenessProbe.Sanitize()
	c.ReadinessProbe.Sanitize()
}

func (c *SidecarContainer) PostProcess() error {
	return multierr.Combine(
		c.Resources.PostProcess(),
		c.LivenessProbe.PostProcess(),
		c.ReadinessProbe.PostProcess(),
	)
}

func (c *SidecarContainer) Validate() error {
	var errs error
	if c.Image == "" {
		errs = multierr.Append(errs, errors.Errorf(".Image must be non-empty"))
	}
	if c.IpFamilyMode != "" && c.IpFamilyMode != "ipv4" && c.IpFamilyMode != "dualstack" {
		errs = multierr.Append(errs, errors.Errorf(".IpFamilyMode must be either 'ipv4' or 'dualstack'"))
	}
	if 65535 < c.RedirectPortInbound {
		errs = multierr.Append(errs, errors.Errorf(".RedirectPortInbound must be in the range [0, 65535]"))
	}
	if 65535 < c.RedirectPortOutbound {
		errs = multierr.Append(errs, errors.Errorf(".RedirectPortOutbound must be in the range [0, 65535]"))
	}
	if 65535 < c.AdminPort {
		errs = multierr.Append(errs, errors.Errorf(".AdminPort must be in the range [0, 65535]"))
	}
	if c.DrainTime.Duration <= 0 {
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
	return errs
}

var _ config.Config = &InitContainer{}

func (c *InitContainer) Validate() error {
	var errs error
	if c.Image == "" {
		errs = multierr.Append(errs, errors.Errorf(".Image must be non-empty"))
	}
	return errs
}

var _ config.Config = &SidecarReadinessProbe{}

func (c *SidecarReadinessProbe) Validate() error {
	var errs error
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
	return errs
}

var _ config.Config = &SidecarLivenessProbe{}

func (c *SidecarLivenessProbe) Validate() error {
	var errs error
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
	return errs
}

var _ config.Config = &SidecarResources{}

func (c *SidecarResources) Sanitize() {
	c.Limits.Sanitize()
	c.Requests.Sanitize()
}

func (c *SidecarResources) PostProcess() error {
	return multierr.Combine(
		c.Limits.PostProcess(),
		c.Requests.PostProcess(),
	)
}

func (c *SidecarResources) Validate() error {
	var errs error
	if err := c.Requests.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Requests is not valid"))
	}
	if err := c.Limits.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Limits is not valid"))
	}
	return errs
}

var _ config.Config = &SidecarResourceRequests{}

func (c *SidecarResourceRequests) Validate() error {
	var errs error
	if _, err := kube_api.ParseQuantity(c.CPU); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".CPU is not valid"))
	}
	if _, err := kube_api.ParseQuantity(c.Memory); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Memory is not valid"))
	}
	return errs
}

var _ config.Config = &SidecarResourceLimits{}

func (c *SidecarResourceLimits) Validate() error {
	var errs error
	if _, err := kube_api.ParseQuantity(c.CPU); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".CPU is not valid"))
	}
	if _, err := kube_api.ParseQuantity(c.Memory); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Memory is not valid"))
	}
	return errs
}

var _ config.Config = &BuiltinDNS{}

func (c *BuiltinDNS) Validate() error {
	var errs error
	if 65535 < c.Port {
		errs = multierr.Append(errs, errors.Errorf(".port must be in the range [0, 65535]"))
	}
	return errs
}
