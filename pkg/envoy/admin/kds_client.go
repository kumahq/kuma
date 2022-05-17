package admin

import (
	"strings"
	"time"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/kds/service"
)

type kdsEnvoyAdminClient struct {
	streams          service.XDSConfigStreams
	operationTimeout time.Duration
}

func NewKDSEnvoyAdminClient(streams service.XDSConfigStreams, operationTimeout time.Duration) EnvoyAdminClient {
	return &kdsEnvoyAdminClient{
		streams:          streams,
		operationTimeout: operationTimeout,
	}
}

var _ EnvoyAdminClient = &kdsEnvoyAdminClient{}

func (k *kdsEnvoyAdminClient) PostQuit(*core_mesh.DataplaneResource) error {
	panic("not implemented")
}

func (k *kdsEnvoyAdminClient) ConfigDump(proxy core_model.ResourceWithAddress) ([]byte, error) {
	zone, nameInZone, err := resNameInZone(proxy.GetMeta().GetName())
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
		return nil, errors.Wrapf(err, "could not send XDSConfigRequest")
	}

	defer k.streams.DeleteWatch(zone, reqId)
	ch := make(chan *mesh_proto.XDSConfigResponse)
	if err := k.streams.WatchResponse(zone, reqId, ch); err != nil {
		return nil, errors.Wrapf(err, "could not watch the response")
	}

	select {
	case <-time.After(k.operationTimeout):
		return nil, errors.Errorf("timeout. Did not receive the response within %v", k.operationTimeout)
	case resp := <-ch:
		if resp.GetError() != "" {
			return nil, errors.Errorf("error response from Zone CP: %s", resp.GetError())
		}
		return resp.GetConfig(), nil
	}
}

func resNameInZone(nameInGlobal string) (string, string, error) {
	parts := strings.Split(nameInGlobal, ".")
	if len(parts) < 2 {
		return "", "", errors.New("wrong name format. Expected {zone}.{name}")
	}
	zone := parts[0] // zone is added by Global CP as a prefix for Dataplane/ZoneIngress/ZoneEgress
	nameInZone := strings.Join(parts[1:], ".")
	return zone, nameInZone, nil
}
