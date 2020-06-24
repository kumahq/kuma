package setup

import (
	"fmt"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	kds_client "github.com/Kong/kuma/pkg/kds/client"
	"github.com/Kong/kuma/pkg/test/grpc"
	. "github.com/onsi/gomega"
)

type mockKDSClient struct {
	kdsStream kds_client.KDSStream
}

func (m *mockKDSClient) StartStream() (kds_client.KDSStream, error) {
	return m.kdsStream, nil
}

func (m *mockKDSClient) Close() error {
	return nil
}

func StartClient(clientStreams []*grpc.MockClientStream, resourceTypes []model.ResourceType, stopCh chan struct{}, cb *kds_client.Callbacks) {
	mgr := component.NewManager()
	fmt.Println("len", len(clientStreams))
	for i := 0; i < len(clientStreams); i++ {
		item := clientStreams[i]
		err := mgr.Add(kds_client.NewKDSSink(core.Log.Logger, resourceTypes, func() (kds_client.KDSClient, error) {
			return &mockKDSClient{kdsStream: kds_client.NewKDSStream(item)}, nil
		}, cb))
		Expect(err).ToNot(HaveOccurred())
	}

	go func() {
		err := mgr.Start(stopCh)
		Expect(err).ToNot(HaveOccurred())
	}()
}
