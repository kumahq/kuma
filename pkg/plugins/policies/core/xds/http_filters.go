package xds

import (
	"errors"

	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
)

func InsertHTTPFiltersBeforeRouter(manager *envoy_hcm.HttpConnectionManager, newFilters ...*envoy_hcm.HttpFilter) error {
	if len(newFilters) == 0 {
		return nil
	}
	for i, filter := range manager.HttpFilters {
		if filter.Name == "envoy.filters.http.router" {
			// insert new filters before router
			manager.HttpFilters = append(manager.HttpFilters[:i+len(newFilters)], manager.HttpFilters[i:]...)
			for x, newFilter := range newFilters {
				manager.HttpFilters[i+x] = newFilter
			}
			return nil
		}
	}
	return errors.New("could not insert filter, envoy.filters.http.router is not found in HTTPConnectionManager")
}
