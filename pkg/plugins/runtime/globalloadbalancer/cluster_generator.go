package globalloadbalancer

import (
	"context"
	"fmt"

	envoy_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/globalloadbalancer/metadata"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
)

type ClusterGenerator struct{}

func (c *ClusterGenerator) GenerateClusters(ctx context.Context, xdsCtx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	// For each Koyeb datacenter with an Ingress Gateway, create a cluster.
	for datacenterID, endpoints := range proxy.GlobalLoadBalancerProxy.EndpointMap {
		clusterName := fmt.Sprintf("dc_%s", datacenterID)

		r, err := generateDatacenterCluster(proxy, clusterName, xdsCtx.Mesh.Resource)
		if err != nil {
			return nil, err
		}
		resources.Add(r)

		r, err = generateLoadAssignment(proxy, clusterName, endpoints)
		if err != nil {
			return nil, err
		}
		resources.Add(r)
	}

	// A Koyeb service can be deployed on 1 to N regions.
	// For each Koyeb service, create an aggregate cluster. The aggregate cluster contains the
	// ordered list of actual clusters where we could redirect incoming requests matched against
	// a Koyeb Service.
	// At index 0 of the aggregate cluster will be the  cluster closest geographically to the
	// current instance of the GLB. At the last index, the geographically furthest away cluster
	// to the current instance of the GLB.
	//
	// For example, if a service `webapp` is deployed in Paris (region `par`, datacenter
	// `par1`) and Washington (region `was1, datacenter `was1`), the global load balancer in
	// Frankfurt will have the following configuration:
	// * cluster 1: aggr_service_webapp{ dc_par1, dc_was1 }
	// * cluster 2: dc_par1 { host=xxxxxx:5601 }
	// * cluster 3: dc_was1 { host=xxxxxx:5601 }
	// The global load balancer in Washington would have the same configuration, except that
	// the aggregate's cluster order would be reversed:
	// * cluster 1: aggr_service_webapp{ dc_was1, dc_par1 }
	// * cluster 2: dc_par1 { host=xxxxxx:5601 }
	// * cluster 3: dc_was1 { host=xxxxxx:5601 }

	for _, app := range proxy.GlobalLoadBalancerProxy.KoyebApps {
		for _, service := range app.Services {
			clusterName := fmt.Sprintf("aggr_service_%s", service.ID)

			r, err := generateAggregateCluster(proxy, clusterName, service)
			if err != nil {
				return nil, err
			}
			resources.Add(r)
		}
	}

	r, err := generateNotFoundCluster(proxy)
	if err != nil {
		return nil, err
	}
	resources.Add(r)

	return resources, nil
}

// This cluster redirects to a 404 page.
// Incoming requests shall be routed there if the Hostname they present is not something
// that we (Koyeb) have knowledge of.
// We want this page to be branded. We cannot simply respon with a 404 and a small text
// because Cloudflare does not allow us to upload a custom 404 page
func generateNotFoundCluster(proxy *core_xds.Proxy) (*core_xds.Resource, error) {
	endpoint := &core_xds.Endpoint{
		Target: "koyeb.github.io",
		Port:   uint32(443),
		ExternalService: &core_xds.ExternalService{
			TLSEnabled:         true,
			ServerName:         "koyeb.github.io",
			AllowRenegotiation: true,
		},
	}

	clusterBuilder := clusters.NewClusterBuilder(proxy.APIVersion, "not_found").Configure(
		clusters.ProvidedEndpointCluster(false, *endpoint),
		clusters.ClientSideTLS([]core_xds.Endpoint{*endpoint}),
	)

	r, err := buildClusterResource(clusterBuilder)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func generateDatacenterCluster(proxy *core_xds.Proxy, clusterName string, mesh *core_mesh.MeshResource) (*core_xds.Resource, error) {
	clusterBuilder := clusters.NewClusterBuilder(proxy.APIVersion, clusterName).Configure(
		clusters.EdsCluster(),
		clusters.UnknownDestinationClientSideMTLS(proxy.SecretsTracker, mesh),
		clusters.ConnectionBufferLimit(DefaultConnectionBuffer),
		clusters.HttpDownstreamProtocolOptions(),
	)

	r, err := buildClusterResource(clusterBuilder)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func buildClusterResource(c *clusters.ClusterBuilder) (*core_xds.Resource, error) {
	msg, err := c.Build()
	if err != nil {
		return nil, err
	}

	cluster := msg.(*envoy_cluster_v3.Cluster)

	return &core_xds.Resource{
		Name:     cluster.GetName(),
		Origin:   metadata.OriginGlobalLoadBalancer,
		Resource: cluster,
	}, nil
}

func generateLoadAssignment(proxy *core_xds.Proxy, clusterName string, endpoints []core_xds.Endpoint) (*core_xds.Resource, error) {
	cla, err := envoy_endpoints.CreateClusterLoadAssignment(clusterName, endpoints, proxy.APIVersion)
	if err != nil {
		return nil, err
	}

	return &core_xds.Resource{
		Name:     clusterName,
		Origin:   metadata.OriginGlobalLoadBalancer,
		Resource: cla,
	}, nil
}

// {"datacenters":[{"id":"was1","region_id":"was","domain":"glb-was1.infra.prod.koyeb.com","coordinates":["-77.007507","38.900497"]},{"id":"tyo1","region_id":"tyo","domain":"glb-tyo1.infra.prod.koyeb.com","coordinates":["139.817413","35.672855"]},{"id":"sin1","region_id":"sin","domain":"glb-sin1.infra.prod.koyeb.com","coordinates":["103.819839","1.352083"]},{"id":"sfo1","region_id":"sfo","domain":"glb-sfo1.infra.prod.koyeb.com","coordinates":["-122.419418","37.774929"]},{"id":"par1","region_id":"par","domain":"glb-par1.infra.prod.koyeb.com","coordinates":["2.3522","48.8566"]},{"id":"fra1","region_id":"fra","domain":"glb-fra1.infra.prod.koyeb.com","coordinates":["8.6821","50.1109"]}]}

func generateAggregateCluster(proxy *core_xds.Proxy, clusterName string, service *core_xds.KoyebService) (*core_xds.Resource, error) {
	datacenters := proxy.GlobalLoadBalancerProxy.Datacenters
	glbDatacenterID := proxy.Dataplane.Spec.TagSet().Values(mesh_proto.KoyebDatacenterTag)[0]
	aggregate, err := GenerateAggregateOfClusters(service, datacenters, glbDatacenterID)
	if err != nil {
		return nil, err
	}

	if len(aggregate) == 0 {
		log.V(1).Info(
			"no available datacenter in the catalog", "serviceId", service.ID,
			"service_datacenters", service.DatacenterIDs, "catalog_datacenters", datacenters,
		)

		return nil, nil
	}

	clusterBuilder := clusters.NewClusterBuilder(proxy.APIVersion, clusterName).Configure(
		clusters.AggregateCluster(aggregate...),
	)

	r, err := buildClusterResource(clusterBuilder)
	if err != nil {
		return nil, err
	}

	return r, nil
}
