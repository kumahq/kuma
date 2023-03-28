package setup

import (
	"fmt"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	kds_client "github.com/kumahq/kuma/pkg/kds/client"
	cache_v2 "github.com/kumahq/kuma/pkg/kds/v2/cache"
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
		item := clientStreams[i]
		comp := kds_client_v2.NewKDSSyncClient(core.Log.WithName("kds").WithName(clientID), resourceTypes, kds_client_v2.NewDeltaKDSStream(item, clientID, "", cache_v2.ResourceVersionMap{}), cb)
		go func() {
			_ = comp.Receive()
			_ = item.CloseSend()
		}()
	}
}
