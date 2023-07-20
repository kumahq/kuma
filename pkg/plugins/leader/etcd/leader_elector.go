package etcd

import (
	"context"
	util_channels "github.com/kumahq/kuma/pkg/util/channels"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"sync/atomic"
	"time"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

var log = core.Log.WithName("etcd-leader")

const (
	kumaLockName = "kuma-cp-lock"
	backoffTime  = 5 * time.Second
)

// etcdLeaderElector implements leader election using etcd.
type etcdLeaderElector struct {
	instanceId string
	leader     int32
	client     *clientv3.Client
	callbacks  []component.LeaderCallbacks
}

var _ component.LeaderElector = &etcdLeaderElector{}

func NewEtcdLeaderElector(instanceId string, client *clientv3.Client) component.LeaderElector {
	return &etcdLeaderElector{
		client:     client,
		instanceId: instanceId,
	}
}

func (e *etcdLeaderElector) Start(stop <-chan struct{}) {
	log.Info("starting Leader Elector")
	ctx, cancelFn := context.WithCancel(context.Background())
	go func() {
		<-stop
		log.Info("stopping Leader Elector")
		if e.IsLeader() {
			e.leaderLost()
		}
		cancelFn()
	}()

	session, err := concurrency.NewSession(e.client)
	if err != nil {
		panic(err)
	}
	election := concurrency.NewElection(session, kumaLockName)

	for {
		if _, err := election.Leader(ctx); err == concurrency.ErrElectionNoLeader {
			if err := election.Campaign(ctx, e.instanceId); err != nil {
				log.Error(err, "etd leader elector error.")
				return
			}
			e.leaderAcquired()
		}
		if util_channels.IsClosed(stop) {
			break
		}
		time.Sleep(backoffTime)
	}

	log.Info("Leader Elector stopped")
}

func (p *etcdLeaderElector) leaderAcquired() {
	p.setLeader(true)
	for _, callback := range p.callbacks {
		callback.OnStartedLeading()
	}
}

func (p *etcdLeaderElector) leaderLost() {
	p.setLeader(false)
	for _, callback := range p.callbacks {
		callback.OnStoppedLeading()
	}
}

func (p *etcdLeaderElector) AddCallbacks(callbacks component.LeaderCallbacks) {
	p.callbacks = append(p.callbacks, callbacks)
}

func (p *etcdLeaderElector) setLeader(leader bool) {
	var value int32 = 0
	if leader {
		value = 1
	}
	atomic.StoreInt32(&p.leader, value)
}

func (p *etcdLeaderElector) IsLeader() bool {
	return atomic.LoadInt32(&(p.leader)) == 1
}
