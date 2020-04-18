package catalog

// Not yet exposed via YAML and env vars on purpose
type CatalogConfig struct {
	ApiServer            ApiServerConfig               `yaml:"-"`
	Bootstrap            BootstrapApiConfig            `yaml:"bootstrap"`
	DataplaneToken       DataplaneTokenApiConfig       `yaml:"-"` // DEPRECATED: remove in next major version of Kuma
	Admin                AdminApiConfig                `yaml:"-"`
	MonitoringAssignment MonitoringAssignmentApiConfig `yaml:"monitoringAssignment"`
	Sds                  SdsApiConfig                  `yaml:"sds"`
}

type ApiServerConfig struct {
	Url string
}

type BootstrapApiConfig struct {
	Url string `yaml:"url" envconfig:"kuma_api_server_catalog_bootstrap_url"`
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
	Url string `yaml:"url" envconfig:"kuma_api_server_catalog_monitoring_assignment_url"`
}

type SdsApiConfig struct {
	Url string `yaml:"url" envconfig:"kuma_api_server_catalog_sds_url"`
}
