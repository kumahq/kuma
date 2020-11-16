package catalog

// Not yet exposed via YAML and env vars on purpose
type CatalogConfig struct {
	Bootstrap      BootstrapApiConfig      `yaml:"bootstrap"` // DEPRECATED: remove in next major version of Kuma
	DataplaneToken DataplaneTokenApiConfig `yaml:"-"`         // DEPRECATED: remove in next major version of Kuma
}

type BootstrapApiConfig struct {
	Url string `yaml:"url" envconfig:"kuma_api_server_catalog_bootstrap_url"`
}

type DataplaneTokenApiConfig struct {
	LocalUrl  string
	PublicUrl string
}

type MonitoringAssignmentApiConfig struct {
	Url string `yaml:"url" envconfig:"kuma_api_server_catalog_monitoring_assignment_url"`
}
