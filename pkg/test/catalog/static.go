package catalog

import (
	"github.com/Kong/kuma/pkg/catalog"
	catalog_client "github.com/Kong/kuma/pkg/catalog/client"
)

type StaticCatalogClient struct {
	Resp catalog.Catalog
}

var _ catalog_client.CatalogClient = &StaticCatalogClient{}

func (s *StaticCatalogClient) Catalog() (catalog.Catalog, error) {
	return s.Resp, nil
}
