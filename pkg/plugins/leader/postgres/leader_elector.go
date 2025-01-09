package postgres

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"sync/atomic"
	"time"

	"cirello.io/pglock"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	util_channels "github.com/kumahq/kuma/pkg/util/channels"
)

var log = core.Log.WithName("postgres-leader")

const (
	kumaLockName = "kuma-cp-lock"
	backoffTime  = 5 * time.Second
)

// postgresLeaderElector implements leader election using PostgreSQL DB.
// pglock does not rely on timestamps, which eliminates the problem of clock skews, but the cost is that first leader election can happen only after lease duration
// pglock does optimistic locking under the hood, the alternative would be to use pg_advisory_lock
type postgresLeaderElector struct {
	leader     int32
	lockClient *pglock.Client
	callbacks  []component.LeaderCallbacks
}

var _ component.LeaderElector = &postgresLeaderElector{}

func NewPostgresLeaderElector(lockClient *pglock.Client) component.LeaderElector {
	return &postgresLeaderElector{
		lockClient: lockClient,
	}
}

func (p *postgresLeaderElector) Start(stop <-chan struct{}) {
	log.Info("starting Leader Elector")
	ctx, cancelFn := context.WithCancel(context.Background())
	go func() {
		<-stop
		log.Info("stopping Leader Elector")
		cancelFn()
	}()

	retries := 0
	for {
		log.Info("waiting for lock")
		err := p.lockClient.Do(ctx, kumaLockName, func(ctx context.Context, lock *pglock.Lock) error {
			p.leaderAcquired()
			<-ctx.Done()
			p.leaderLost()
			return nil
		})
		// in case of error (ex. connection to postgres is dropped) we want to retry the lock with some backoff
		// returning error here would shut down the CP
		// according to https://github.com/cirello-io/pglock/blob/d0d1ce72df710b5da6bdc27f9c44d9ae7bf1d3a2/errors.go#L86
		// ErrNotAcquired could be normal lock contestation
		if err != nil {
			if retries >= 3 {
				log.Error(err, "error waiting for lock", "retries", retries)
			} else {
				log.V(1).Info("error waiting for lock", "err", err, "retries", retries)
			}
			retries += 1
		} else {
			retries = 0
		}

		if util_channels.IsClosed(stop) {
			break
		}
		time.Sleep(backoffTime)
	}
	log.Info("Leader Elector stopped")
}

func (p *postgresLeaderElector) leaderAcquired() {
	p.setLeader(true)
	for _, callback := range p.callbacks {
		callback.OnStartedLeading()
	}
}

func (p *postgresLeaderElector) leaderLost() {
	p.setLeader(false)
	for _, callback := range p.callbacks {
		callback.OnStoppedLeading()
	}
}

func (p *postgresLeaderElector) AddCallbacks(callbacks component.LeaderCallbacks) {
	p.callbacks = append(p.callbacks, callbacks)
}

func (p *postgresLeaderElector) setLeader(leader bool) {
	var value int32 = 0
	if leader {
		value = 1
	}
	atomic.StoreInt32(&p.leader, value)
}

func (p *postgresLeaderElector) IsLeader() bool {
	return atomic.LoadInt32(&(p.leader)) == 1
}

type KumaPqLockLogger struct{}

func (k *KumaPqLockLogger) Error(msg string, args ...interface{}) {
	if len(args) > 0 {
		err, ok := args[0].(error)
		if ok {
			if !errors.Is(err, context.Canceled) {
				log.Error(err, fmt.Sprintf(msg, args...))
			}
		}
	} else {
		log.Error(nil, fmt.Sprintf(msg, args...))
	}
}

func (k *KumaPqLockLogger) Debug(msg string, args ...interface{}) {
	log.V(1).Info(fmt.Sprintf(msg, args...))
}
