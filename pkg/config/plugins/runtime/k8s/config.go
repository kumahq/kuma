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
		Injector: Injector{
			CNIEnabled:           false,
			VirtualProbesEnabled: true,
			VirtualProbesPort:    9000,
			SidecarContainer: SidecarContainer{
				Image:                "kuma/kuma-dp:latest",
				RedirectPortInbound:  15006,
				RedirectPortOutbound: 15001,
				UID:                  5678,
				GID:                  5678,
				AdminPort:            9901,
				DrainTime:            30 * time.Second,

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
			InitContainer: InitContainer{
				Image: "kuma/kuma-init:latest",
			},
		},
	}
}

// Kubernetes-specific configuration
type KubernetesRuntimeConfig struct {
	// Admission WebHook Server implemented by the Control Plane.
	AdmissionServer AdmissionServerConfig `yaml:"admissionServer"`
	// Injector-specific configuration
	Injector Injector `yaml:"injector,omitempty"`
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
	// CNIEnabled if true runs kuma-cp in CNI compatible mode
	CNIEnabled bool `yaml:"cniEnabled" envconfig:"kuma_runtime_kubernetes_injector_cni_enabled"`
	// VirtualProbesEnabled enables automatic converting HttpGet probes to virtual. Virtual probe
	// serves on sub-path of insecure port 'virtualProbesPort',
	// i.e :8080/health/readiness -> :9000/8080/health/readiness where 9000 is virtualProbesPort
	VirtualProbesEnabled bool `yaml:"virtualProbesEnabled" envconfig:"kuma_runtime_kubernetes_virtual_probes_enabled"`
	// VirtualProbesPort is an insecure port for listening virtual probes
	VirtualProbesPort uint32 `yaml:"virtualProbesPort" envconfig:"kuma_runtime_kubernetes_virtual_probes_enabled"`
}

// SidecarContainer defines configuration of the Kuma sidecar container.
type SidecarContainer struct {
	// Image name.
	Image string `yaml:"image,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_image"`
	// Redirect port for inbound traffic.
	RedirectPortInbound uint32 `yaml:"redirectPortInbound,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_redirect_port_inbound"`
	// Redirect port for outbound traffic.
	RedirectPortOutbound uint32 `yaml:"redirectPortOutbound,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_redirect_port_outbound"`
	// User ID.
	UID int64 `yaml:"uid,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_uid"`
	// Group ID.
	GID int64 `yaml:"gid,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_gui"`
	// Admin port.
	AdminPort uint32 `yaml:"adminPort,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_admin_port"`
	// Drain time for listeners.
	DrainTime time.Duration `yaml:"drainTime,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_drain_time"`
	// Readiness probe.
	ReadinessProbe SidecarReadinessProbe `yaml:"readinessProbe,omitempty"`
	// Liveness probe.
	LivenessProbe SidecarLivenessProbe `yaml:"livenessProbe,omitempty"`
	// Compute resource requirements.
	Resources SidecarResources `yaml:"resources,omitempty"`
}

// SidecarReadinessProbe defines periodic probe of container service readiness.
type SidecarReadinessProbe struct {
	// Number of seconds after the container has started before liveness probes are initiated.
	InitialDelaySeconds int32 `yaml:"initialDelaySeconds,omitempty" envconfig:"kuma_runtime_kubernetes_injector_sidecar_container_readiness_probe_initial_delay_seconds"`
	// How often (in seconds) to perform the probe.
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
