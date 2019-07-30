package konvoyinjector

import (
	"net"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
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
				XdsServer: XdsServer{
					Address: "konvoy-control-plane.konvoy-system",
					Port:    5678,
				},
				ApiServer: ApiServer{
					Address: "konvoy-control-plane.konvoy-system",
					Port:    5681,
				},
			},
			SidecarContainer: SidecarContainer{
				Image:        "envoyproxy/envoy-alpine:latest",
				RedirectPort: 15001,
				UID:          5678,
				GID:          5678,
				AdminPort:    9901,
			},
			InitContainer: InitContainer{
				Image: "docker.io/istio/proxy_init:1.1.2",
			},
		},
	}
}

// Config defines configuration of the Konvoy Injector.
type Config struct {
	// WebHookServer defines configuration of an https server that implements Kubernetes Admission WebHook.
	WebHookServer WebHookServer `yaml:"webHookServer,omitempty"`
	// PodTemplate defines configuration of the Konvoy Sidecar Injector.
	Injector Injector `yaml:"injector,omitempty"`
}

// WebHookServer defines configuration of an https server that implements Kubernetes Admission WebHook.
type WebHookServer struct {
	// Address defines the address the https server should be listening on.
	Address string `yaml:"address,omitempty" envconfig:"konvoy_injector_webhook_server_address"`
	// Port defines the port the https server should be listening on.
	Port uint32 `yaml:"port,omitempty" envconfig:"konvoy_injector_webhook_server_port"`
	// CertDir defines path to a directory with TLS certificate and key for the https server.
	// TLS certificate file must be named `tls.crt`.
	// TLS key file must be named `tls.key`.
	CertDir string `yaml:"certDir,omitempty" envconfig:"konvoy_injector_webhook_server_cert_dir"`
}

// Injector defines configuration of a Konvoy Sidecar Injector.
type Injector struct {
	// ControlPlane defines coordinates of the Konvoy Control Plane.
	ControlPlane ControlPlane `yaml:"controlPlane,omitempty"`
	// SidecarContainer defines configuration of the Konvoy sidecar container.
	SidecarContainer SidecarContainer `yaml:"sidecarContainer,omitempty"`
	// InitContainer defines configuration of the Konvoy init container.
	InitContainer InitContainer `yaml:"initContainer,omitempty"`
}

// ControlPlane defines coordinates of the Konvoy Control Plane.
type ControlPlane struct {
	// XdsServer defines coordinates of the Konvoy xDS Server.
	XdsServer XdsServer `yaml:"xdsServer,omitempty"`
	// ApiServer defines coordinates of the Konvoy API Server.
	ApiServer ApiServer `yaml:"apiServer,omitempty"`
}

// XdsServer defines coordinates of the Konvoy xDS Server.
type XdsServer struct {
	Address string `yaml:"address,omitempty" envconfig:"konvoy_injector_control_plane_xds_server_address"`
	Port    uint32 `yaml:"port,omitempty" envconfig:"konvoy_injector_control_plane_xds_server_port"`
}

// ApiServer defines coordinates of the Konvoy API Server.
type ApiServer struct {
	Address string `yaml:"address,omitempty" envconfig:"konvoy_injector_control_plane_api_server_address"`
	Port    uint32 `yaml:"port,omitempty" envconfig:"konvoy_injector_control_plane_api_server_port"`
}

// SidecarContainer defines configuration of the Konvoy sidecar container.
type SidecarContainer struct {
	// Image name.
	Image string `yaml:"image,omitempty" envconfig:"konvoy_injector_sidecar_container_image"`
	// Redirect port.
	RedirectPort uint32 `yaml:"redirectPort,omitempty" envconfig:"konvoy_injector_sidecar_container_redirect_port"`
	// User ID.
	UID int64 `yaml:"uid,omitempty" envconfig:"konvoy_injector_sidecar_container_uid"`
	// Group ID.
	GID int64 `yaml:"gid,omitempty" envconfig:"konvoy_injector_sidecar_container_gui"`
	// Admin port.
	AdminPort uint32 `yaml:"adminPort,omitempty" envconfig:"konvoy_injector_sidecar_container_admin_port"`
}

// InitContainer defines configuration of the Konvoy init container.
type InitContainer struct {
	// Image name.
	Image string `yaml:"image,omitempty" envconfig:"konvoy_injector_init_container_image"`
}

var _ config.Config = &Config{}

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

func (c *ControlPlane) Validate() (errs error) {
	if err := c.XdsServer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".XdsServer is not valid"))
	}
	if err := c.ApiServer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".ApiServer is not valid"))
	}
	return
}

var _ config.Config = &XdsServer{}

func (s *XdsServer) Validate() (errs error) {
	if s.Address == "" {
		errs = multierr.Append(errs, errors.Errorf(".Address must be non-empty"))
	}
	if 65535 < s.Port {
		errs = multierr.Append(errs, errors.Errorf(".Port must be in the range [0, 65535]"))
	}
	return
}

var _ config.Config = &ApiServer{}

func (s *ApiServer) Validate() (errs error) {
	if s.Address == "" {
		errs = multierr.Append(errs, errors.Errorf(".Address must be non-empty"))
	}
	if 65535 < s.Port {
		errs = multierr.Append(errs, errors.Errorf(".Port must be in the range [0, 65535]"))
	}
	return
}

var _ config.Config = &SidecarContainer{}

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
	return
}

var _ config.Config = &InitContainer{}

func (c *InitContainer) Validate() (errs error) {
	if c.Image == "" {
		errs = multierr.Append(errs, errors.Errorf(".Image must be non-empty"))
	}
	return
}
