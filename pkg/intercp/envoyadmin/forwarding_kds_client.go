package envoyadmin

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/intercp/catalog"
)

var clientLog = core.Log.WithName("intercp").WithName("envoyadmin").WithName("client")

type NewClientFn = func(url string) (mesh_proto.InterCPEnvoyAdminForwardServiceClient, error)

type forwardingKdsEnvoyAdminClient struct {
	resManager     manager.ReadOnlyResourceManager
	cat            catalog.Catalog
	instanceID     string
	newClientFn    NewClientFn
	fallbackClient admin.EnvoyAdminClient
}

// NewForwardingEnvoyAdminClient returns EnvoyAdminClient which is only used on Global CP in multizone environment.
// It forwards the request to an instance of the Global CP to which Zone CP of given DPP is connected.
//
// For example:
// We have 2 instances of Global CP (ins-1, ins-2). Dataplane "backend" is in zone "east".
// The leader CP of zone "east" is connected to ins-1.
// If we execute config dump for "backend" on ins-1, we follow the regular flow of pkg/envoy/admin/kds_client.go
// If we execute config dump for "backend" on ins-2, we forward the request to ins-1 and then execute the regular flow.
func NewForwardingEnvoyAdminClient(
	resManager manager.ReadOnlyResourceManager,
	cat catalog.Catalog,
	instanceID string,
	newClientFn NewClientFn,
	fallbackClient admin.EnvoyAdminClient,
) admin.EnvoyAdminClient {
	return &forwardingKdsEnvoyAdminClient{
		resManager:     resManager,
		cat:            cat,
		instanceID:     instanceID,
		newClientFn:    newClientFn,
		fallbackClient: fallbackClient,
	}
}

var _ admin.EnvoyAdminClient = &forwardingKdsEnvoyAdminClient{}

func (f *forwardingKdsEnvoyAdminClient) PostQuit(context.Context, *core_mesh.DataplaneResource) error {
	panic("not implemented")
}

func (f *forwardingKdsEnvoyAdminClient) ConfigDump(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
	instanceID, err := f.globalInstanceID(ctx, core_model.ZoneOfResource(proxy))
	if err != nil {
		return nil, err
	}
	f.logIntendedAction(proxy, instanceID)
	if instanceID == f.instanceID {
		return f.fallbackClient.ConfigDump(ctx, proxy)
	}
	client, err := f.clientForInstanceID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	req := &mesh_proto.XDSConfigRequest{
		ResourceType: string(proxy.Descriptor().Name),
		ResourceName: proxy.GetMeta().GetName(),
		ResourceMesh: proxy.GetMeta().GetMesh(),
	}
	resp, err := client.XDSConfig(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetConfig(), nil
}

func (f *forwardingKdsEnvoyAdminClient) Stats(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
	instanceID, err := f.globalInstanceID(ctx, core_model.ZoneOfResource(proxy))
	if err != nil {
		return nil, err
	}
	f.logIntendedAction(proxy, instanceID)
	if instanceID == f.instanceID {
		return f.fallbackClient.Stats(ctx, proxy)
	}
	client, err := f.clientForInstanceID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	req := &mesh_proto.StatsRequest{
		ResourceType: string(proxy.Descriptor().Name),
		ResourceName: proxy.GetMeta().GetName(),
		ResourceMesh: proxy.GetMeta().GetMesh(),
	}
	resp, err := client.Stats(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetStats(), nil
}

func (f *forwardingKdsEnvoyAdminClient) Clusters(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
	instanceID, err := f.globalInstanceID(ctx, core_model.ZoneOfResource(proxy))
	if err != nil {
		return nil, err
	}
	f.logIntendedAction(proxy, instanceID)
	if instanceID == f.instanceID {
		return f.fallbackClient.Clusters(ctx, proxy)
	}
	client, err := f.clientForInstanceID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	req := &mesh_proto.ClustersRequest{
		ResourceType: string(proxy.Descriptor().Name),
		ResourceName: proxy.GetMeta().GetName(),
		ResourceMesh: proxy.GetMeta().GetMesh(),
	}
	resp, err := client.Clusters(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetClusters(), nil
}

func (f *forwardingKdsEnvoyAdminClient) logIntendedAction(proxy core_model.ResourceWithAddress, instanceID string) {
	log := clientLog.WithValues(
		"name", proxy.GetMeta().GetName(),
		"mesh", proxy.GetMeta().GetMesh(),
		"type", proxy.Descriptor().Name,
		"instanceID", instanceID,
	)
	if instanceID == f.instanceID {
		log.V(1).Info("zone CP of the resource is connected to this Global CP instance. Executing operation")
	} else {
		log.V(1).Info("zone CP of the resource is connected to other Global CP instance. Forwarding the request")
	}
}

func (f *forwardingKdsEnvoyAdminClient) globalInstanceID(ctx context.Context, zone string) (string, error) {
	zoneInsightRes := core_system.NewZoneInsightResource()
	if err := f.resManager.Get(ctx, zoneInsightRes, core_store.GetByKey(zone, core_model.NoMesh)); err != nil {
		return "", err
	}
	sub := zoneInsightRes.Spec.GetLastSubscription().(*system_proto.KDSSubscription)
	return sub.GlobalInstanceId, nil
}

func (f *forwardingKdsEnvoyAdminClient) clientForInstanceID(ctx context.Context, instanceID string) (mesh_proto.InterCPEnvoyAdminForwardServiceClient, error) {
	instance, err := catalog.InstanceOfID(ctx, f.cat, instanceID)
	if err != nil {
		return nil, err
	}
	return f.newClientFn(instance.InterCpURL())
}
