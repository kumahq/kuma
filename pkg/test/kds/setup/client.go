package setup

import (
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	kds_client "github.com/kumahq/kuma/pkg/kds/client"
	"github.com/kumahq/kuma/pkg/test/grpc"
)

func StartClient(clientStreams []*grpc.MockClientStream, resourceTypes []model.ResourceType, stopCh chan struct{}, cb *kds_client.Callbacks) {
	for i := 0; i < len(clientStreams); i++ {
		item := clientStreams[i]
		comp := kds_client.NewKDSSink(core.Log.Logger, resourceTypes, kds_client.NewKDSStream(item, "client-1", ""), cb)
		go func() {
			_ = comp.Start(stopCh)
			_ = item.CloseSend()
		}()
	}
}
