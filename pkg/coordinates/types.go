package coordinates

import (
	"fmt"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
)

type Coordinates struct {
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

func FromConfig(cfg kuma_cp.Config) Coordinates {
	result := Coordinates{
		Apis: Apis{
			Bootstrap: BootstrapApi{
				Url: fmt.Sprintf("http://%s:%d", cfg.Hostname, cfg.BootstrapServer.Port),
			},
			DataplaneToken: DataplaneTokenApi{
				LocalUrl: fmt.Sprintf("http://localhost:%d", cfg.DataplaneTokenServer.Local.Port),
			},
		},
	}

	if cfg.DataplaneTokenServer.TlsEnabled() {
		result.Apis.DataplaneToken.PublicUrl = fmt.Sprintf("https://%s:%d", cfg.Hostname, cfg.DataplaneTokenServer.Public.Port)
	}
	return result
}
