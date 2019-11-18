package catalog

// Not yet exposed via YAML and env vars on purpose
type CatalogConfig struct {
	Bootstrap      BootstrapApiConfig
	DataplaneToken DataplaneTokenApiConfig
}

type BootstrapApiConfig struct {
	Url string
}

type DataplaneTokenApiConfig struct {
	LocalUrl  string
	PublicUrl string
}
