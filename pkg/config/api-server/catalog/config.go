package catalog

// Not yet exposed via YAML and env vars on purpose
type CatalogConfig struct {
	Bootstrap            BootstrapApiConfig            `yaml:"bootstrap"`
	DataplaneToken       DataplaneTokenApiConfig       `yaml:"-"` // DEPRECATED: remove in next major version of Kuma
	Admin                AdminApiConfig                `yaml:"-"`
	MonitoringAssignment MonitoringAssignmentApiConfig `yaml:"monitoringAssignment"`
	Sds                  SdsApiConfig                  `yaml:"sds"`
}

type BootstrapApiConfig struct {
	Url string `yaml:"url" envconfig:"kuma_bootstrap_server_public_url"`
}

type DataplaneTokenApiConfig struct {
	LocalUrl  string
	PublicUrl string
}

type AdminApiConfig struct {
	LocalUrl  string
	PublicUrl string
}

type MonitoringAssignmentApiConfig struct {
	Url string `yaml:"url" envconfig:"kuma_mads_server_public_url"`
}

type SdsApiConfig struct {
	HostPort string `yaml:"hostPort" envconfig:"kuma_sds_server_public_host_port"`
}
