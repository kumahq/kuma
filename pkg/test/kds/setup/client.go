package setup

import (
	"fmt"

	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/v2/pkg/core/runtime"
	kds_client_v2 "github.com/kumahq/kuma/v2/pkg/kds/v2/client"
	"github.com/kumahq/kuma/v2/pkg/test/grpc"
)

func StartDeltaClient(clientStreams []*grpc.MockDeltaClientStream, resourceTypes []model.ResourceType, stopCh chan struct{}, cb *kds_client_v2.Callbacks) {
	for i := 0; i < len(clientStreams); i++ {
		clientID := fmt.Sprintf("client-%d", i)
		runtimeInfo := core_runtime.NewRuntimeInfo(fmt.Sprintf("cp-%d", i), config_core.Zone)
		item := clientStreams[i]
<<<<<<< HEAD
		comp := kds_client_v2.NewKDSSyncClient(core.Log.WithName("kds").WithName(clientID), resourceTypes, kds_client_v2.NewDeltaKDSStream(item, clientID, runtimeInfo, "", len(resourceTypes)), cb, 0)
=======
		kdsStream := kds_client_v2.NewDeltaKDSStream(item, clientID, fmt.Sprintf("cp-%d", i), "", len(resourceTypes))
		comp := kds_client_v2.NewKDSSyncClient(core.Log.WithName("kds").WithName(clientID), resourceTypes, kdsStream, cb, 0)
>>>>>>> 666d45dc0f (fix(kds): reconnect mux client when GlobalToZone stream is closed by … (#16326))
		go func() {
			_ = comp.Receive()
			_ = kdsStream.CloseSend()
		}()
	}
}
