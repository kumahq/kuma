package catalog

import (
	"github.com/kumahq/kuma/pkg/catalog"
	catalog_client "github.com/kumahq/kuma/pkg/catalog/client"
)

type StaticCatalogClient struct {
	Resp catalog.Catalog
}

var _ catalog_client.CatalogClient = &StaticCatalogClient{}

func (s *StaticCatalogClient) Catalog() (catalog.Catalog, error) {
	return s.Resp, nil
}
