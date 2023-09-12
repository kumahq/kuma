package sync

import (
	"context"
	"net"
	"sort"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/coord"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type GlobalLoadBalancerProxyBuilder struct {
	*DataplaneProxyBuilder
}

func (p *GlobalLoadBalancerProxyBuilder) Build(ctx context.Context, key core_model.ResourceKey, meshContext xds_context.MeshContext) (*core_xds.Proxy, error) {
	proxy, err := p.DataplaneProxyBuilder.Build(ctx, key, meshContext)
	if err != nil {
		return nil, err
	}

	// TODO(nicoche) fetch this from the API
	parCoord, err := coord.NewCoord([]string{"2.3522", "48.8566"})
	if err != nil {
		return nil, err
	}

	datacenters := []*core_xds.KoyebDatacenter{
		{
			ID:       "par1",
			RegionID: "par",
			Coord:    parCoord,
			Domain:   "localhost",
		},
	}

	endpointMap, err := p.buildEndpointMap(datacenters)
	if err != nil {
		return nil, err
	}

	koyebApps := p.fetchKoyebApps()

	proxy.GlobalLoadBalancerProxy = &core_xds.GlobalLoadBalancerProxy{
		Datacenters: datacenters,
		EndpointMap: endpointMap,
		KoyebApps:   koyebApps,
	}

	return proxy, nil
}

func (p *GlobalLoadBalancerProxyBuilder) buildEndpointMap(datacenters []*core_xds.KoyebDatacenter) (core_xds.EndpointMap, error) {
	endpointMap := core_xds.EndpointMap{}

	for _, dc := range datacenters {
		// TODO(nicoche): move that in a goroutine
		ips, err := net.LookupIP(dc.Domain)
		if err != nil {
			return nil, errors.Wrapf(err, "could not resolve %s", dc.Domain)
		}

		// Not sure that we need it, but always generate the slice of endpoints in
		// the same order.
		sort.Slice(ips, func(i, j int) bool {
			return ips[i].String() < ips[j].String()
		})

		for _, ip := range ips {
			endpointMap[dc.ID] = append(endpointMap[dc.ID], core_xds.Endpoint{
				Target: ip.String(),
				// TODO(nicoche) should we harmonize that? urgh it's ugly
				Port:   5602,
				Weight: 1,
			})
		}
	}

	return endpointMap, nil
}

func (p *GlobalLoadBalancerProxyBuilder) fetchKoyebApps() []*core_xds.KoyebApp {
	return []*core_xds.KoyebApp{
		{
			Domains: []string{"grpc.koyeb.app"},
			Services: []*core_xds.KoyebService{
				{
					ID: "dp",
					DatacenterIDs: map[string]struct{}{
						"par1": {},
					},
					Port:            8004,
					DeploymentGroup: "prod",
					Paths:           []string{""},
				},
			},
		},
		{
			Domains: []string{"http.local.koyeb.app"},
			Services: []*core_xds.KoyebService{
				{
					ID: "dp",
					DatacenterIDs: map[string]struct{}{
						"par1": {},
					},
					Port:            8001,
					DeploymentGroup: "prod",
					Paths:           []string{"/http"},
				},
				{
					ID: "dp",
					DatacenterIDs: map[string]struct{}{
						"par1": {},
					},
					Port:            8002,
					DeploymentGroup: "prod",
					Paths:           []string{"/http2"},
				},
				{
					ID: "dp",
					DatacenterIDs: map[string]struct{}{
						"par1": {},
					},
					Port:            8004,
					DeploymentGroup: "prod",
					Paths:           []string{"/grpc"},
				},
				{
					ID: "dp",
					DatacenterIDs: map[string]struct{}{
						"par1": {},
					},
					Port:            8011,
					DeploymentGroup: "prod",
					Paths:           []string{"/ws"},
				},
			},
		},
	}
}
