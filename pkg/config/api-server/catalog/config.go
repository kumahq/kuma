package catalog

// Not yet exposed via YAML and env vars on purpose
type CatalogConfig struct {
	Bootstrap      BootstrapApiConfig
	DataplaneToken DataplaneTokenApiConfig // DEPRECATED: remove in next major version of Kuma
	Admin          AdminApiConfig
}

type BootstrapApiConfig struct {
	Url string
}

type DataplaneTokenApiConfig struct {
	LocalUrl  string
	PublicUrl string
}

type AdminApiConfig struct {
	LocalUrl  string
	PublicUrl string
}
