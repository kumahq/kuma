package setup

import (
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/model"
	kds_client "github.com/Kong/kuma/pkg/kds/client"
	"github.com/Kong/kuma/pkg/test/grpc"
)

type mockKDSClient struct {
	kdsStream kds_client.KDSStream
}

func (m *mockKDSClient) StartStream(clientId string) (kds_client.KDSStream, error) {
	return m.kdsStream, nil
}

func (m *mockKDSClient) Close() error {
	return nil
}

func StartClient(clientStreams []*grpc.MockClientStream, resourceTypes []model.ResourceType, stopCh chan struct{}, cb *kds_client.Callbacks) {
	for i := 0; i < len(clientStreams); i++ {
		item := clientStreams[i]
		comp := kds_client.NewKDSSink(core.Log.Logger, "global", resourceTypes, func() (kds_client.KDSClient, error) {
			return &mockKDSClient{kdsStream: kds_client.NewKDSStream(item, "client-1")}, nil
		}, cb)
		go func() {
			_ = comp.Start(stopCh)
		}()
	}
}
