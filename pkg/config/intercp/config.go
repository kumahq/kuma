package intercp

import (
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	config_types "github.com/kumahq/kuma/pkg/config/types"
)

func DefaultInterCpConfig() InterCpConfig {
	return InterCpConfig{
		Catalog: CatalogConfig{
			InstanceAddress:   "", // autoconfigured
			HeartbeatInterval: config_types.Duration{Duration: 5 * time.Second},
			WriterInterval:    config_types.Duration{Duration: 15 * time.Second},
		},
		Server: InterCpServerConfig{
			Port:            5683,
			TlsMinVersion:   "TLSv1_2",
			TlsCipherSuites: []string{},
		},
	}
}

type InterCpConfig struct {
	// Catalog configuration. Catalog keeps a record of all live CP instances in the zone.
	Catalog CatalogConfig `json:"catalog"`
	// Intercommunication CP server configuration
	Server InterCpServerConfig `json:"server"`
}

func (i *InterCpConfig) Validate() error {
	if err := i.Server.Validate(); err != nil {
		return errors.Wrap(err, ".Server validation failed")
	}
	if err := i.Catalog.Validate(); err != nil {
		return errors.Wrap(err, ".Catalog validation failed")
	}
	return nil
}

type CatalogConfig struct {
	// InstanceAddress indicates an address on which other control planes can communicate with this CP
	// If empty then it's autoconfigured by taking the first IP of the nonloopback network interface.
	InstanceAddress string `json:"instanceAddress" envconfig:"kuma_inter_cp_catalog_instance_address"`
	// Interval on which CP will send heartbeat to a leader.
	HeartbeatInterval config_types.Duration `json:"heartbeatInterval" envconfig:"kuma_inter_cp_catalog_heartbeat_interval"`
	// Interval on which CP will write all instances to a catalog.
	WriterInterval config_types.Duration `json:"writerInterval" envconfig:"kuma_inter_cp_catalog_writer_interval"`
}

func (i *CatalogConfig) Validate() error {
	if i.InstanceAddress != "" && !govalidator.IsDNSName(i.InstanceAddress) && !govalidator.IsIP(i.InstanceAddress) {
		return errors.New(".InstanceAddress has to be valid IP or DNS address")
	}
	return nil
}

type InterCpServerConfig struct {
	// Port on which Intercommunication CP server will listen
	Port uint16 `json:"port" envconfig:"kuma_inter_cp_server_port"`
	// TlsMinVersion defines the minimum TLS version to be used
	TlsMinVersion string `json:"tlsMinVersion" envconfig:"kuma_inter_cp_server_tls_min_version"`
	// TlsMaxVersion defines the maximum TLS version to be used
	TlsMaxVersion string `json:"tlsMaxVersion" envconfig:"kuma_inter_cp_server_tls_max_version"`
	// TlsCipherSuites defines the list of ciphers to use
	TlsCipherSuites []string `json:"tlsCipherSuites" envconfig:"kuma_inter_cp_server_tls_cipher_suites"`
}

func (i *InterCpServerConfig) Validate() error {
	var errs error
	if i.Port == 0 {
		errs = multierr.Append(errs, errors.New(".Port cannot be zero"))
	}
	if _, err := config_types.TLSVersion(i.TlsMinVersion); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsMinVersion "+err.Error()))
	}
	if _, err := config_types.TLSVersion(i.TlsMaxVersion); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsMaxVersion "+err.Error()))
	}
	if _, err := config_types.TLSCiphers(i.TlsCipherSuites); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsCipherSuites "+err.Error()))
	}
	return errs
}
