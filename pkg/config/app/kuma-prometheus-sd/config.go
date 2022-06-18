package kuma_prometheus_sd

import (
	"fmt"
	"net/url"

	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/mads"
)

var DefaultConfig = func() Config {
	return Config{
		MonitoringAssignment: MonitoringAssignmentConfig{
			Client: MonitoringAssignmentClientConfig{
				Name:       "kuma_sd",
				URL:        "grpc://localhost:5676",
				ApiVersion: mads.API_V1,
			},
		},
		Prometheus: PrometheusConfig{
			OutputFile: "kuma.file_sd.json",
		},
	}
}

// Config defines configuration of the Prometheus service discovery adapter.
type Config struct {
	// MonitoringAssignment defines configuration related to Monitoring Assignment in Kuma.
	MonitoringAssignment MonitoringAssignmentConfig `yaml:"monitoringAssignment,omitempty"`
	// Prometheus defines configuration related to integration with Prometheus.
	Prometheus PrometheusConfig `yaml:"prometheus,omitempty"`
}

var _ config.Config = &Config{}

func (c *Config) Sanitize() {
	c.MonitoringAssignment.Sanitize()
	c.Prometheus.Sanitize()
}

func (c *Config) Validate() (errs error) {
	if err := c.MonitoringAssignment.Validate(); err != nil {
		errs = multierr.Append(errs, fmt.Errorf(".MonitoringAssignment is not valid: %w", err))
	}
	if err := c.Prometheus.Validate(); err != nil {
		errs = multierr.Append(errs, fmt.Errorf(".Prometheus is not valid: %w", err))
	}
	return
}

// MonitoringAssignmentConfig defines configuration related to Monitoring Assignment in Kuma.
type MonitoringAssignmentConfig struct {
	// Client defines configuration of a Monitoring Assignment Discovery Service (MADS) client.
	Client MonitoringAssignmentClientConfig `yaml:"client,omitempty"`
}

// MonitoringAssignmentClientConfig defines configuration of a
// Monitoring Assignment Discovery Service (MADS) client.
type MonitoringAssignmentClientConfig struct {
	// Address defines the address of Control Plane's Monitoring Assignment Discovery Service server.
	URL string `yaml:"url,omitempty" envconfig:"kuma_monitoring_assignment_client_url"`
	// Name this adapter should use when connecting to Monitoring Assignment server.
	Name string `yaml:"name,omitempty" envconfig:"kuma_monitoring_assignment_client_name"`
	// ApiVersion is the MADS API version served by the Monitoring Assignment server.
	ApiVersion string `yaml:"apiVersion,omitempty" envconfig:"kuma_monitoring_assignment_client_api_version"`
}

var _ config.Config = &MonitoringAssignmentConfig{}

func (c *MonitoringAssignmentConfig) Sanitize() {
}

func (c *MonitoringAssignmentConfig) Validate() (errs error) {
	if err := c.Client.Validate(); err != nil {
		errs = multierr.Append(errs, fmt.Errorf(".Client is not valid: %w", err))
	}
	return
}

var _ config.Config = &MonitoringAssignmentClientConfig{}

func (c *MonitoringAssignmentClientConfig) Sanitize() {
}

func (c *MonitoringAssignmentClientConfig) Validate() (errs error) {
	if c.Name == "" {
		errs = multierr.Append(errs, fmt.Errorf(".Name must be non-empty"))
	}
	if c.URL == "" {
		errs = multierr.Append(errs, fmt.Errorf(".URL must be non-empty"))
	}
	url, err := url.Parse(c.URL)
	if err != nil {
		errs = multierr.Append(errs, fmt.Errorf(".URL must be a valid absolute URI: %w", err))
	} else {
		if !url.IsAbs() {
			errs = multierr.Append(errs, fmt.Errorf(".URL must be a valid absolute URI"))
		}
		if url.Scheme != "grpc" && url.Scheme != "grpcs" {
			errs = multierr.Append(errs, fmt.Errorf(".URL must start with grpc:// or grpcs://"))
		}
	}

	if c.ApiVersion == "" {
		errs = multierr.Append(errs, fmt.Errorf(".ApiVersion must be non-empty"))
	} else if c.ApiVersion != mads.API_V1 {
		errs = multierr.Append(errs, fmt.Errorf(".ApiVersion must be v1, got: %s", c.ApiVersion))
	}

	return
}

// PrometheusConfig defines configuration related to integration with Prometheus.
//
// In short, Kuma Prometheus SD adapter integrates with Prometheus via a shared file,
// where the former is a writer and the latter is a reader.
// For further details see https://github.com/prometheus/prometheus/tree/master/documentation/examples/custom-sd
type PrometheusConfig struct {
	// Path to an output file where Kuma Prometheus SD adapter should persists a list of scrape targets.
	// The same file path must be used on Prometheus side in a configuration of `file_sd` discovery mechanism.
	OutputFile string `yaml:"outputFile,omitempty" envconfig:"kuma_prometheus_output_file"`
}

var _ config.Config = &PrometheusConfig{}

func (c *PrometheusConfig) Sanitize() {
}

func (c *PrometheusConfig) Validate() (errs error) {
	if c.OutputFile == "" {
		errs = multierr.Append(errs, fmt.Errorf(".OutputFile must be non-empty"))
	}
	return
}
