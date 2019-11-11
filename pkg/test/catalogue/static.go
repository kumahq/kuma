package catalogue

import (
	"github.com/Kong/kuma/pkg/catalogue"
	catalogue_client "github.com/Kong/kuma/pkg/catalogue/client"
)

type StaticCatalogueClient struct {
	Resp catalogue.Catalogue
}

var _ catalogue_client.CatalogueClient = &StaticCatalogueClient{}

func (s *StaticCatalogueClient) Catalogue() (catalogue.Catalogue, error) {
	return s.Resp, nil
}
