package setup

import (
	"fmt"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	kds_client_v2 "github.com/kumahq/kuma/pkg/kds/v2/client"
	"github.com/kumahq/kuma/pkg/test/grpc"
)

func StartDeltaClient(clientStreams []*grpc.MockDeltaClientStream, resourceTypes []model.ResourceType, stopCh chan struct{}, cb *kds_client_v2.Callbacks) {
	for i := 0; i < len(clientStreams); i++ {
		clientID := fmt.Sprintf("client-%d", i)
		runtimeInfo := core_runtime.NewRuntimeInfo(fmt.Sprintf("cp-%d", i), config_core.Zone)
		item := clientStreams[i]
<<<<<<< HEAD
		comp := kds_client_v2.NewKDSSyncClient(core.Log.WithName("kds").WithName(clientID), resourceTypes, kds_client_v2.NewDeltaKDSStream(item, clientID, runtimeInfo, ""), cb, 0)
=======
		comp := kds_client_v2.NewKDSSyncClient(core.Log.WithName("kds").WithName(clientID), resourceTypes, kds_client_v2.NewDeltaKDSStream(item, clientID, fmt.Sprintf("cp-%d", i), "", len(resourceTypes)), cb, 0)
>>>>>>> c4f7db2534 (fix(kds): server Send blocks when client doesn't call Recv for some time (#15042))
		go func() {
			_ = comp.Receive()
			_ = item.CloseSend()
		}()
	}
}
