package catalog

import (
	"github.com/kumahq/kuma/pkg/config/api-server/catalog"
)

type Catalog struct {
	Apis Apis `json:"apis"`
}

type Apis struct {
	Bootstrap      BootstrapApi      `json:"bootstrap"`
	DataplaneToken DataplaneTokenApi `json:"dataplaneToken"` // DEPRECATED: remove in next major version of Kuma
}

type AdminApi struct {
	LocalUrl  string `json:"localUrl"`
	PublicUrl string `json:"publicUrl"`
}

type BootstrapApi struct {
	Url string `json:"url"`
}

type DataplaneTokenApi struct {
	LocalUrl  string `json:"localUrl"`
	PublicUrl string `json:"publicUrl"`
}

type MonitoringAssignmentApi struct {
	Url string `json:"url"`
}

func (d *DataplaneTokenApi) Enabled() bool {
	return d.LocalUrl != ""
}

func FromConfig(cfg catalog.CatalogConfig) Catalog {
	return Catalog{
		Apis: Apis{
			Bootstrap: BootstrapApi{
				Url: cfg.Bootstrap.Url,
			},
			DataplaneToken: DataplaneTokenApi{
				LocalUrl:  cfg.DataplaneToken.LocalUrl,
				PublicUrl: cfg.DataplaneToken.PublicUrl,
			},
		},
	}
}
