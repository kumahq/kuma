package setup

import (
	"fmt"
	"time"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	kds_client "github.com/kumahq/kuma/pkg/kds/client"
	kds_client_v2 "github.com/kumahq/kuma/pkg/kds/v2/client"
	"github.com/kumahq/kuma/pkg/test/grpc"
)

func StartClient(clientStreams []*grpc.MockClientStream, resourceTypes []model.ResourceType, stopCh chan struct{}, cb *kds_client.Callbacks) {
	for i := 0; i < len(clientStreams); i++ {
		clientID := fmt.Sprintf("client-%d", i)
		item := clientStreams[i]
		comp := kds_client.NewKDSSink(core.Log.WithName("kds").WithName(clientID), resourceTypes, kds_client.NewKDSStream(item, clientID, ""), cb)
		go func() {
			_ = comp.Receive()
			_ = item.CloseSend()
		}()
	}
}

func StartDeltaClient(clientStreams []*grpc.MockDeltaClientStream, resourceTypes []model.ResourceType, stopCh chan struct{}, cb *kds_client_v2.Callbacks) {
	for i := 0; i < len(clientStreams); i++ {
		clientID := fmt.Sprintf("client-%d", i)
		runtimeInfo := &mockRuntimeInfo{
			instanceId: fmt.Sprintf("cp-%d", i),
			mode:       config_core.Zone,
		}
		item := clientStreams[i]
		comp := kds_client_v2.NewKDSSyncClient(core.Log.WithName("kds").WithName(clientID), resourceTypes, kds_client_v2.NewDeltaKDSStream(item, clientID, runtimeInfo, ""), cb, 0)
		go func() {
			_ = comp.Receive()
			_ = item.CloseSend()
		}()
	}
}

type mockRuntimeInfo struct {
	instanceId string
	mode       config_core.CpMode
}

func (m mockRuntimeInfo) GetInstanceId() string {
	return m.instanceId
}

func (m mockRuntimeInfo) SetClusterId(clusterId string) {
	panic("implement me")
}

func (m mockRuntimeInfo) GetClusterId() string {
	panic("implement me")
}

func (m mockRuntimeInfo) GetStartTime() time.Time {
	panic("implement me")
}

func (m mockRuntimeInfo) GetMode() config_core.CpMode {
	return m.mode
}
