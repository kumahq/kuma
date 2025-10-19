package setup

import (
	"context"
	"fmt"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	kds_client_v2 "github.com/kumahq/kuma/pkg/kds/v2/client"
	kds_util "github.com/kumahq/kuma/pkg/kds/v2/util"
	"github.com/kumahq/kuma/pkg/test/grpc"
	"golang.org/x/sync/errgroup"
)

func StartDeltaClient(clientStreams []*grpc.MockDeltaClientStream, resourceTypes []model.ResourceType, stopCh chan struct{}, cb *kds_util.Callbacks) {
	for i := 0; i < len(clientStreams); i++ {
		clientID := fmt.Sprintf("client-%d", i)
		item := clientStreams[i]
		comp := kds_client_v2.NewKDSSyncClient(core.Log.WithName("kds").WithName(clientID), resourceTypes, kds_client_v2.NewDeltaKDSStream(item, clientID, fmt.Sprintf("cp-%d", i), ""), cb, 0)
		group, ctx := errgroup.WithContext(context.TODO())
		go func() {
			_ = comp.Receive(ctx, group)
			_ = item.CloseSend()
		}()
	}
}
