package event

import (
	"context"
	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/plugins/common/etcd"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var log = core.Log.WithName("etcd-event-listener")

type listener struct {
	prefix string
	client clientv3.Client
	out    events.Emitter
}

func (l listener) Start(stop <-chan struct{}) error {
	log.Info("etcd listener starting...")
	ctx := context.Background()
	prefixKey := etcd.NewEtcdKey(l.prefix, "", "", "").Prefix()

	watchChan := l.client.Watch(ctx, prefixKey, clientv3.WithPrefix())
	for {
		select {
		case <-stop:
			log.Info("etcd listener stoping...")
			return nil
		case resp := <-watchChan:
			for _, event := range resp.Events {
				var op events.Op
				switch event.Type {
				case mvccpb.DELETE:
					op = events.Delete
				case mvccpb.PUT:
					op = events.Update
				}
				key := event.Kv.Key

				etcdKey, err := etcd.WithEtcdKey(string(key))
				if err != nil {
					log.Error(err, "key", key)
				}

				l.out.Send(events.ResourceChangedEvent{
					Operation: op,
					Type:      etcdKey.GetResourceType(),
					Key:       core_model.ResourceKey{Mesh: etcdKey.GetMesh(), Name: etcdKey.GetName()},
					TenantID:  "",
				})
			}
		}
	}

	return nil
}

func (l listener) NeedLeaderElection() bool {
	return false
}

func NewListener(prefix string, client clientv3.Client, out events.Emitter) *listener {
	return &listener{prefix: prefix, client: client, out: out}
}
