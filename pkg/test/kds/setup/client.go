package setup

import (
	"fmt"

	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	kds_client "github.com/kumahq/kuma/v3/pkg/kds/client"
	"github.com/kumahq/kuma/v3/pkg/test/grpc"
)

// StartDeltaClient starts a KDS sync client per stream. clientIDs[i] is the
// client-id for clientStreams[i] (the zone name in production, which drives
// attribution) and must be provided for every stream.
func StartDeltaClient(clientStreams []*grpc.MockDeltaClientStream, clientIDs []string, resourceTypes []model.ResourceType, stopCh chan struct{}, cb *kds_client.Callbacks) {
	for i := range clientStreams {
		clientID := clientIDs[i]
		item := clientStreams[i]
		kdsStream := kds_client.NewDeltaKDSStream(item, clientID, fmt.Sprintf("cp-%d", i), "", len(resourceTypes))
		comp := kds_client.NewKDSSyncClient(
			core.Log.WithName("kds").WithName(clientID),
			resourceTypes,
			kdsStream,
			cb,
			kds_client.SyncClientConfig{},
		)
		go func() {
			_ = comp.Receive()
			_ = kdsStream.CloseSend()
		}()
	}
}
