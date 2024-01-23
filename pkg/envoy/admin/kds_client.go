package admin

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/kds/service"
	util_grpc "github.com/kumahq/kuma/pkg/util/grpc"
	"github.com/kumahq/kuma/pkg/util/k8s"
)

type kdsEnvoyAdminClient struct {
	rpcs service.EnvoyAdminRPCs
}

func NewKDSEnvoyAdminClient(rpcs service.EnvoyAdminRPCs) EnvoyAdminClient {
	return &kdsEnvoyAdminClient{
		rpcs: rpcs,
	}
}

var _ EnvoyAdminClient = &kdsEnvoyAdminClient{}

func (k *kdsEnvoyAdminClient) PostQuit(context.Context, *core_mesh.DataplaneResource) error {
	panic("not implemented")
}

func (k *kdsEnvoyAdminClient) ConfigDump(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
	zone := core_model.ZoneOfResource(proxy)
	nameInZone := resNameInZone(proxy)
	reqId := core.NewUUID()
	tenantZoneID := service.TenantZoneClientIDFromCtx(ctx, zone)

	err := k.rpcs.XDSConfigDump.Send(tenantZoneID.String(), &mesh_proto.XDSConfigRequest{
		RequestId:    reqId,
		ResourceType: string(proxy.Descriptor().Name),
		ResourceName: nameInZone,                // send the name which without the added prefix
		ResourceMesh: proxy.GetMeta().GetMesh(), // should be empty for ZoneIngress/ZoneEgress
	})
	if err != nil {
		return nil, &KDSTransportError{requestType: "XDSConfigRequest", reason: err.Error()}
	}

	defer k.rpcs.XDSConfigDump.DeleteWatch(tenantZoneID.String(), reqId)
	ch := make(chan util_grpc.ReverseUnaryMessage)
	if err := k.rpcs.XDSConfigDump.WatchResponse(tenantZoneID.String(), reqId, ch); err != nil {
		return nil, errors.Wrapf(err, "could not watch the response")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-ch:
		configResp, ok := resp.(*mesh_proto.XDSConfigResponse)
		if !ok {
			return nil, errors.New("invalid request type")
		}
		if configResp.GetError() != "" {
			return nil, &KDSTransportError{requestType: "XDSConfigRequest", reason: configResp.GetError()}
		}
		return configResp.GetConfig(), nil
	}
}

func (k *kdsEnvoyAdminClient) Stats(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
	zone := core_model.ZoneOfResource(proxy)
	nameInZone := resNameInZone(proxy)
	reqId := core.NewUUID()
	tenantZoneId := service.TenantZoneClientIDFromCtx(ctx, zone)

	err := k.rpcs.Stats.Send(tenantZoneId.String(), &mesh_proto.StatsRequest{
		RequestId:    reqId,
		ResourceType: string(proxy.Descriptor().Name),
		ResourceName: nameInZone,                // send the name which without the added prefix
		ResourceMesh: proxy.GetMeta().GetMesh(), // should be empty for ZoneIngress/ZoneEgress
	})
	if err != nil {
		return nil, &KDSTransportError{requestType: "StatsRequest", reason: err.Error()}
	}

	defer k.rpcs.Stats.DeleteWatch(tenantZoneId.String(), reqId)
	ch := make(chan util_grpc.ReverseUnaryMessage)
	if err := k.rpcs.Stats.WatchResponse(tenantZoneId.String(), reqId, ch); err != nil {
		return nil, errors.Wrapf(err, "could not watch the response")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-ch:
		statsResp, ok := resp.(*mesh_proto.StatsResponse)
		if !ok {
			return nil, errors.New("invalid request type")
		}
		if statsResp.GetError() != "" {
			return nil, &KDSTransportError{requestType: "StatsRequest", reason: statsResp.GetError()}
		}
		return statsResp.GetStats(), nil
	}
}

func (k *kdsEnvoyAdminClient) Clusters(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
	zone := core_model.ZoneOfResource(proxy)
	nameInZone := resNameInZone(proxy)
	reqId := core.NewUUID()
	tenantZoneID := service.TenantZoneClientIDFromCtx(ctx, zone)

	err := k.rpcs.Clusters.Send(tenantZoneID.String(), &mesh_proto.ClustersRequest{
		RequestId:    reqId,
		ResourceType: string(proxy.Descriptor().Name),
		ResourceName: nameInZone,                // send the name which without the added prefix
		ResourceMesh: proxy.GetMeta().GetMesh(), // should be empty for ZoneIngress/ZoneEgress
	})
	if err != nil {
		return nil, &KDSTransportError{requestType: "ClustersRequest", reason: err.Error()}
	}

	defer k.rpcs.Clusters.DeleteWatch(tenantZoneID.String(), reqId)
	ch := make(chan util_grpc.ReverseUnaryMessage)
	if err := k.rpcs.Clusters.WatchResponse(tenantZoneID.String(), reqId, ch); err != nil {
		return nil, errors.Wrapf(err, "could not watch the response")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-ch:
		clustersResp, ok := resp.(*mesh_proto.ClustersResponse)
		if !ok {
			return nil, errors.New("invalid request type")
		}
		if clustersResp.GetError() != "" {
			return nil, &KDSTransportError{requestType: "ClustersRequest", reason: clustersResp.GetError()}
		}
		return clustersResp.GetClusters(), nil
	}
}

func resNameInZone(r core_model.Resource) string {
	name := core_model.GetDisplayName(r)
	// we need to check for the legacy name which starts with zoneName
	if strings.HasPrefix(r.GetMeta().GetName(), core_model.ZoneOfResource(r)) {
		return name
	}
	// since 2.6 zone sets store type so we can figure out if we need to add namespace
	if core_model.GetOriginStoreType(r) != store.KubernetesStore {
		return name
	}
	if ns := r.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag]; ns != "" {
		return k8s.K8sNamespacedNameToCoreName(name, ns)
	}
	return name
}

type KDSTransportError struct {
	requestType string
	reason      string
}

func (e *KDSTransportError) Error() string {
	if e.reason == "" {
		return fmt.Sprintf("could not send %s", e.requestType)
	} else {
		return fmt.Sprintf("could not send %s: %s", e.requestType, e.reason)
	}
}

func (e *KDSTransportError) Is(err error) bool {
	return reflect.TypeOf(e) == reflect.TypeOf(err)
}
