package kumainjector

import (
	"net"
	"net/url"
	"time"

	"github.com/Kong/kuma/pkg/config"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	kube_api "k8s.io/apimachinery/pkg/api/resource"
)

func DefaultConfig() Config {
	return Config{
		WebHookServer: WebHookServer{
			// By default, https server will be listening on all addresses.
			Address: "",
			// By default, https server will be listening on port 8443.
			Port: 8443,
			// CertDir has no default value and must always be set explicitly.
			CertDir: "",
		},
		Injector: Injector{
			ControlPlane: ControlPlane{
				ApiServer: ApiServer{
					URL: "http://kuma-control-plane.kuma-system:5681",
				},
			},
			SidecarContainer: SidecarContainer{
				Image:        "kuma/kuma-dp:latest",
				RedirectPort: 15001,
				UID:          5678,
				GID:          5678,
				AdminPort:    9901,
				DrainTime:    30 * time.Second,

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
				Image: "docker.io/istio/proxy_init:1.1.2",
			},
		},
	}
}

// Config defines configuration of the Kuma Injector.
type Config struct {
	// WebHookServer defines configuration of an https server that implements Kubernetes Admission WebHook.
	WebHookServer WebHookServer `yaml:"webHookServer,omitempty"`
	// PodTemplate defines configuration of the Kuma Sidecar Injector.
	Injector Injector `yaml:"injector,omitempty"`
}

// WebHookServer defines configuration of an https server that implements Kubernetes Admission WebHook.
type WebHookServer struct {
	// Address defines the address the https server should be listening on.
	Address string `yaml:"address,omitempty" envconfig:"kuma_injector_webhook_server_address"`
	// Port defines the port the https server should be listening on.
	Port uint32 `yaml:"port,omitempty" envconfig:"kuma_injector_webhook_server_port"`
	// CertDir defines path to a directory with TLS certificate and key for the https server.
	// TLS certificate file must be named `tls.crt`.
	// TLS key file must be named `tls.key`.
	CertDir string `yaml:"certDir,omitempty" envconfig:"kuma_injector_webhook_server_cert_dir"`
}

func (s *WebHookServer) Sanitize() {
}

// Injector defines configuration of a Kuma Sidecar Injector.
type Injector struct {
	// ControlPlane defines coordinates of the Kuma Control Plane.
	ControlPlane ControlPlane `yaml:"controlPlane,omitempty"`
	// SidecarContainer defines configuration of the Kuma sidecar container.
	SidecarContainer SidecarContainer `yaml:"sidecarContainer,omitempty"`
	// InitContainer defines configuration of the Kuma init container.
	InitContainer InitContainer `yaml:"initContainer,omitempty"`
}

// ControlPlane defines coordinates of the Control Plane.
type ControlPlane struct {
	// ApiServer defines coordinates of the Control Plane API Server.
	ApiServer ApiServer `yaml:"apiServer,omitempty"`
}

// ApiServer defines coordinates of the Control Plane API Server.
type ApiServer struct {
	// URL defines URL of the Control Plane API Server.
	URL string `yaml:"url,omitempty" envconfig:"kuma_injector_control_plane_api_server_url"`
}

// SidecarContainer defines configuration of the Kuma sidecar container.
type SidecarContainer struct {
	// Image name.
	Image string `yaml:"image,omitempty" envconfig:"kuma_injector_sidecar_container_image"`
	// Redirect port.
	RedirectPort uint32 `yaml:"redirectPort,omitempty" envconfig:"kuma_injector_sidecar_container_redirect_port"`
	// User ID.
	UID int64 `yaml:"uid,omitempty" envconfig:"kuma_injector_sidecar_container_uid"`
	// Group ID.
	GID int64 `yaml:"gid,omitempty" envconfig:"kuma_injector_sidecar_container_gui"`
	// Admin port.
	AdminPort uint32 `yaml:"adminPort,omitempty" envconfig:"kuma_injector_sidecar_container_admin_port"`
	// Drain time for listeners.
	DrainTime time.Duration `yaml:"drainTime,omitempty" envconfig:"kuma_injector_sidecar_container_drain_time"`
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
	InitialDelaySeconds int32 `yaml:"initialDelaySeconds,omitempty" envconfig:"kuma_injector_sidecar_container_readiness_probe_initial_delay_seconds"`
	// How often (in seconds) to perform the probe.
	TimeoutSeconds int32 `yaml:"timeoutSeconds,omitempty" envconfig:"kuma_injector_sidecar_container_readiness_probe_timeout_seconds"`
	// Number of seconds after which the probe times out.
	PeriodSeconds int32 `yaml:"periodSeconds,omitempty" envconfig:"kuma_injector_sidecar_container_readiness_probe_period_seconds"`
	// Minimum consecutive successes for the probe to be considered successful after having failed.
	SuccessThreshold int32 `yaml:"successThreshold,omitempty" envconfig:"kuma_injector_sidecar_container_readiness_probe_success_threshold"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	FailureThreshold int32 `yaml:"failureThreshold,omitempty" envconfig:"kuma_injector_sidecar_container_readiness_probe_failure_threshold"`
}

// SidecarLivenessProbe defines periodic probe of container service liveness.
type SidecarLivenessProbe struct {
	// Number of seconds after the container has started before liveness probes are initiated.
	InitialDelaySeconds int32 `yaml:"initialDelaySeconds,omitempty" envconfig:"kuma_injector_sidecar_container_liveness_probe_initial_delay_seconds"`
	// Number of seconds after which the probe times out.
	TimeoutSeconds int32 `yaml:"timeoutSeconds,omitempty" envconfig:"kuma_injector_sidecar_container_liveness_probe_timeout_seconds"`
	// How often (in seconds) to perform the probe.
	PeriodSeconds int32 `yaml:"periodSeconds,omitempty" envconfig:"kuma_injector_sidecar_container_liveness_probe_period_seconds"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	FailureThreshold int32 `yaml:"failureThreshold,omitempty" envconfig:"kuma_injector_sidecar_container_liveness_probe_failure_threshold"`
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

var _ config.Config = &Config{}

func (c *Config) Sanitize() {
	c.Injector.Sanitize()
	c.WebHookServer.Sanitize()
}

func (c *Config) Validate() (errs error) {
	if err := c.WebHookServer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".WebHookServer is not valid"))
	}
	if err := c.Injector.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Injector is not valid"))
	}
	return
}

var _ config.Config = &WebHookServer{}

func (s *WebHookServer) Validate() (errs error) {
	if s.Address != "" && net.ParseIP(s.Address) == nil {
		errs = multierr.Append(errs, errors.Errorf(".Address must be either empty or a valid IPv4/IPv6 address"))
	}
	if 65535 < s.Port {
		errs = multierr.Append(errs, errors.Errorf(".Port must be in the range [0, 65535]"))
	}
	if s.CertDir == "" {
		errs = multierr.Append(errs, errors.Errorf(".CertDir must be non-empty"))
	}
	return
}

var _ config.Config = &Injector{}

func (i *Injector) Sanitize() {
	i.ControlPlane.Sanitize()
	i.InitContainer.Sanitize()
	i.SidecarContainer.Sanitize()
}

func (i *Injector) Validate() (errs error) {
	if err := i.ControlPlane.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".ControlPlane is not valid"))
	}
	if err := i.SidecarContainer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".SidecarContainer is not valid"))
	}
	if err := i.InitContainer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".InitContainer is not valid"))
	}
	return
}

var _ config.Config = &ControlPlane{}

func (c *ControlPlane) Sanitize() {
	c.ApiServer.Sanitize()
}

func (c *ControlPlane) Validate() (errs error) {
	if err := c.ApiServer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".ApiServer is not valid"))
	}
	return
}

var _ config.Config = &ApiServer{}

func (s *ApiServer) Sanitize() {
}

func (s *ApiServer) Validate() (errs error) {
	if s.URL == "" {
		errs = multierr.Append(errs, errors.Errorf(".URL must be non-empty"))
	}
	if url, err := url.Parse(s.URL); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".URL must be a valid absolute URI"))
	} else if !url.IsAbs() {
		errs = multierr.Append(errs, errors.Errorf(".URL must be a valid absolute URI"))
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
	if 65535 < c.RedirectPort {
		errs = multierr.Append(errs, errors.Errorf(".RedirectPort must be in the range [0, 65535]"))
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
