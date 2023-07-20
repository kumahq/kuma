package watchdog

import (
	"context"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdWatchdog struct {
	client clientv3.Client
	node   *envoy_core.Node
}

func (e *EtcdWatchdog) Start(stop <-chan struct{}) {
	key := e.node.Id
	ctx := context.Background()
	watchChan := e.client.Watch(ctx, key, clientv3.WithPrefix())

	go func() {
		for {
			select {
			case <-stop:
				e.client.Close()
			case resp := <-watchChan:
				for _, event := range resp.Events {
					event = event
				}

			}
		}
	}()
}
