package globalloadbalancer_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kumahq/kuma/pkg/coord"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/globalloadbalancer"
)

func newCoord(t *testing.T, c []string) coord.Coord {
	ret, err := coord.NewCoord(c)
	require.NoError(t, err)
	return ret
}

func TestGenerateAggregateOfClusters(t *testing.T) {
	is := require.New(t)

	// Given
	datacenters := []*core_xds.KoyebDatacenter{
		{
			ID:    "par1",
			Coord: newCoord(t, []string{"48.8566", "2.3522"}),
		},
		{
			ID:    "nyc1",
			Coord: newCoord(t, []string{"40.7128", "74.0060"}),
		},
		{
			ID:    "bxl1",
			Coord: newCoord(t, []string{"50.8503", "4.3517"}),
		},
		{
			ID:    "lon1",
			Coord: newCoord(t, []string{"51.5072", "0.1276"}),
		},
	}
	service := &core_xds.KoyebService{
		DatacenterIDs: map[string]struct{}{
			"bxl1": {},
			"nyc1": {},
		},
	}

	// When
	aggregate, err := globalloadbalancer.GenerateAggregateOfClusters(service, datacenters, "par1")
	// Then
	is.NoError(err)
	is.Equal([]string{"dc_bxl1", "dc_nyc1"}, aggregate)

	// When
	aggregate, err = globalloadbalancer.GenerateAggregateOfClusters(service, datacenters, "nyc1")
	// Then
	is.NoError(err)
	is.Equal([]string{"dc_nyc1", "dc_bxl1"}, aggregate)
}
