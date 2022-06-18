package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/kds/service"
)

type kdsEnvoyAdminClient struct {
	streams  service.XDSConfigStreams
	k8sStore bool
}

func NewKDSEnvoyAdminClient(streams service.XDSConfigStreams, k8sStore bool) EnvoyAdminClient {
	return &kdsEnvoyAdminClient{
		streams:  streams,
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
	err = k.streams.Send(zone, &mesh_proto.XDSConfigRequest{
		RequestId:    reqId,
		ResourceType: string(proxy.Descriptor().Name),
		ResourceName: nameInZone,                // send the name which without the added prefix
		ResourceMesh: proxy.GetMeta().GetMesh(), // should be empty for ZoneIngress/ZoneEgress
	})
	if err != nil {
		return nil, fmt.Errorf("could not send XDSConfigRequest: %w", err)
	}

	defer k.streams.DeleteWatch(zone, reqId)
	ch := make(chan *mesh_proto.XDSConfigResponse)
	if err := k.streams.WatchResponse(zone, reqId, ch); err != nil {
		return nil, fmt.Errorf("could not watch the response: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, errors.New("timeout")
	case resp := <-ch:
		if resp.GetError() != "" {
			return nil, fmt.Errorf("error response from Zone CP: %s", resp.GetError())
		}
		return resp.GetConfig(), nil
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
