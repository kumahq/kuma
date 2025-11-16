package setup

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	kds_client_v2 "github.com/kumahq/kuma/v2/pkg/kds/v2/client"
	kds_util "github.com/kumahq/kuma/v2/pkg/kds/v2/util"
	"github.com/kumahq/kuma/v2/pkg/test/grpc"
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
