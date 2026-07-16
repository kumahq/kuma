package setup

import (
	"fmt"

	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	kds_client_v2 "github.com/kumahq/kuma/v3/pkg/kds/v2/client"
	"github.com/kumahq/kuma/v3/pkg/test/grpc"
)

// StartDeltaClient starts a KDS sync client per stream. clientIDs[i] is the
// authenticated peer identity for clientStreams[i]; in production this is the
// zone name from util.ClientIDFromIncomingCtx and it drives resource
// attribution on the global ingest path. If clientIDs is shorter than
// clientStreams the remaining streams fall back to "client-<i>".
func StartDeltaClient(clientStreams []*grpc.MockDeltaClientStream, clientIDs []string, resourceTypes []model.ResourceType, stopCh chan struct{}, cb *kds_client_v2.Callbacks) {
	for i := range clientStreams {
		clientID := fmt.Sprintf("client-%d", i)
		if i < len(clientIDs) {
			clientID = clientIDs[i]
		}
		item := clientStreams[i]
		kdsStream := kds_client_v2.NewDeltaKDSStream(item, clientID, fmt.Sprintf("cp-%d", i), "", len(resourceTypes))
		comp := kds_client_v2.NewKDSSyncClient(
			core.Log.WithName("kds").WithName(clientID),
			resourceTypes,
			kdsStream,
			cb,
			kds_client_v2.SyncClientConfig{},
		)
		go func() {
			_ = comp.Receive()
			_ = kdsStream.CloseSend()
		}()
	}
}
