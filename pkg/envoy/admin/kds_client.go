package admin

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/kds/service"
	util_grpc "github.com/kumahq/kuma/pkg/util/grpc"
)

type kdsEnvoyAdminClient struct {
	rpcs     service.EnvoyAdminRPCs
	k8sStore bool
}

func NewKDSEnvoyAdminClient(rpcs service.EnvoyAdminRPCs, k8sStore bool) EnvoyAdminClient {
	return &kdsEnvoyAdminClient{
		rpcs:     rpcs,
		k8sStore: k8sStore,
	}
}

var _ EnvoyAdminClient = &kdsEnvoyAdminClient{}

func (k *kdsEnvoyAdminClient) PostQuit(context.Context, *core_mesh.DataplaneResource) error {
	panic("not implemented")
}

func (k *kdsEnvoyAdminClient) ConfigDump(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
	zone, nameInZone, err := resNameInZone(proxy.GetMeta().GetName(), k.k8sStore)
	if err != nil {
		return nil, err
	}
	reqId := core.NewUUID()
	clientID := service.ClientID(ctx, zone)
	err = k.rpcs.XDSConfigDump.Send(clientID, &mesh_proto.XDSConfigRequest{
		RequestId:    reqId,
		ResourceType: string(proxy.Descriptor().Name),
		ResourceName: nameInZone,                // send the name which without the added prefix
		ResourceMesh: proxy.GetMeta().GetMesh(), // should be empty for ZoneIngress/ZoneEgress
	})
	if err != nil {
		return nil, errors.Wrapf(err, "could not send XDSConfigRequest")
	}

	defer k.rpcs.XDSConfigDump.DeleteWatch(clientID, reqId)
	ch := make(chan util_grpc.ReverseUnaryMessage)
	if err := k.rpcs.XDSConfigDump.WatchResponse(clientID, reqId, ch); err != nil {
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
			return nil, errors.Errorf("error response from Zone CP: %s", configResp.GetError())
		}
		return configResp.GetConfig(), nil
	}
}

func (k *kdsEnvoyAdminClient) Stats(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
	zone, nameInZone, err := resNameInZone(proxy.GetMeta().GetName(), k.k8sStore)
	if err != nil {
		return nil, err
	}
	reqId := core.NewUUID()
	clientID := service.ClientID(ctx, zone)
	err = k.rpcs.Stats.Send(clientID, &mesh_proto.StatsRequest{
		RequestId:    reqId,
		ResourceType: string(proxy.Descriptor().Name),
		ResourceName: nameInZone,                // send the name which without the added prefix
		ResourceMesh: proxy.GetMeta().GetMesh(), // should be empty for ZoneIngress/ZoneEgress
	})
	if err != nil {
		return nil, errors.Wrapf(err, "could not send StatsRequest")
	}

	defer k.rpcs.Stats.DeleteWatch(clientID, reqId)
	ch := make(chan util_grpc.ReverseUnaryMessage)
	if err := k.rpcs.Stats.WatchResponse(clientID, reqId, ch); err != nil {
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
			return nil, errors.Errorf("error response from Zone CP: %s", statsResp.GetError())
		}
		return statsResp.GetStats(), nil
	}
}

func (k *kdsEnvoyAdminClient) Clusters(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
	zone, nameInZone, err := resNameInZone(proxy.GetMeta().GetName(), k.k8sStore)
	if err != nil {
		return nil, err
	}
	reqId := core.NewUUID()
	clientID := service.ClientID(ctx, zone)
	err = k.rpcs.Clusters.Send(clientID, &mesh_proto.ClustersRequest{
		RequestId:    reqId,
		ResourceType: string(proxy.Descriptor().Name),
		ResourceName: nameInZone,                // send the name which without the added prefix
		ResourceMesh: proxy.GetMeta().GetMesh(), // should be empty for ZoneIngress/ZoneEgress
	})
	if err != nil {
		return nil, errors.Wrapf(err, "could not send ClustersRequest")
	}

	defer k.rpcs.Clusters.DeleteWatch(clientID, reqId)
	ch := make(chan util_grpc.ReverseUnaryMessage)
	if err := k.rpcs.Clusters.WatchResponse(clientID, reqId, ch); err != nil {
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
			return nil, errors.Errorf("error response from Zone CP: %s", clustersResp.GetError())
		}
		return clustersResp.GetClusters(), nil
	}
}

func resNameInZone(nameInGlobal string, k8sStore bool) (string, string, error) {
	parts := strings.Split(nameInGlobal, ".")
	if len(parts) < 2 {
		return "", "", errors.New("wrong name format. Expected {zone}.{name}")
	}
	zone := parts[0] // zone is added by Global CP as a prefix for Dataplane/ZoneIngress/ZoneEgress
	var nameInZone string
	if k8sStore {
		// if the type of store is Kubernetes then DPP resources are stored in namespaces. Kuma core model
		// is not aware of namespaces and that's why the name in the core model is equal to 'name + .<namespace>'.
		// Before sending the request we should trim the namespace suffix
		nameInZone = strings.Join(parts[1:len(parts)-1], ".")
	} else {
		nameInZone = strings.Join(parts[1:], ".")
	}
	return zone, nameInZone, nil
}
