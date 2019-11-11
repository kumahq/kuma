package catalogue

import (
	"github.com/Kong/kuma/pkg/config/api-server/catalogue"
)

type Catalogue struct {
	Apis Apis `json:"apis"`
}

type Apis struct {
	Bootstrap      BootstrapApi      `json:"bootstrap"`
	DataplaneToken DataplaneTokenApi `json:"dataplaneToken"`
}

type BootstrapApi struct {
	Url string `json:"url"`
}

type DataplaneTokenApi struct {
	LocalUrl  string `json:"localUrl"`
	PublicUrl string `json:"publicUrl"`
}

func (d *DataplaneTokenApi) Enabled() bool {
	return d.LocalUrl != ""
}

func FromConfig(cfg catalogue.CatalogueConfig) Catalogue {
	return Catalogue{
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
