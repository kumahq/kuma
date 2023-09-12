package globalloadbalancer

import (
	"fmt"
	"sort"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

// GenerateAggregateOfClusters returns the expected slice of clusters to fuel the
// service's aggregate cluster with.
//
// Given a service with dcs par1, bxl1, ny1, if the current code runs on bxl1, we expect to
// have an aggregate of [dc_bxl1, dc_par1, dc_ny1].
// This makes dc_bxl1 the primary cluster to route to, dc_par1 the secondary and dc_ny1 the
// third one. We want to always have the primary to be the closest to the current DC, for
// obvious latency reasons.
func GenerateAggregateOfClusters(service *core_xds.KoyebService, datacenters []*core_xds.KoyebDatacenter, glbDatacenterID string) ([]string, error) {
	aggregate := []string{}

	datacenter, err := findDatacenter(datacenters, glbDatacenterID)
	if err != nil {
		return nil, err
	}

	// Sort our slice of datacenters by closest from the GLB to the furthest away.
	sort.SliceStable(datacenters, func(i, j int) bool {
		return datacenters[i].Coord.Distance(datacenter.Coord) < datacenters[j].Coord.Distance(datacenter.Coord)
	})

	// Iterate over all datacenters, from closest to furthest away.
	for _, dc := range datacenters {
		_, ok := service.DatacenterIDs[dc.ID]
		// If the key access works, it means that the service is deployed in that
		// datacenter.
		// Let's add it to our list of targets to route to.
		if ok {
			aggregate = append(aggregate, fmt.Sprintf("dc_%s", dc.ID))
		}
	}

	return aggregate, nil
}

func findDatacenter(datacenters []*core_xds.KoyebDatacenter, id string) (*core_xds.KoyebDatacenter, error) {
	for _, datacenter := range datacenters {
		if datacenter.ID == id {
			return datacenter, nil
		}
	}

	return nil, fmt.Errorf("could not find the GLB's datacenters %s in the available DCs list %v", id, datacenters)
}
