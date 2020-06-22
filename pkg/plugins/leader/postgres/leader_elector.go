package postgres

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"cirello.io/pglock"

	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	util_channels "github.com/Kong/kuma/pkg/util/channels"
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
	ctx, cancelFn := context.WithCancel(context.Background())
	go func() {
		<-stop
		log.Info("Stopping Leader Elector")
		cancelFn()
	}()

	for {
		log.Info("Waiting for lock")
		err := p.lockClient.Do(ctx, kumaLockName, func(ctx context.Context, lock *pglock.Lock) error {
			p.setLeader(true)
			for _, callback := range p.callbacks {
				callback.OnStartedLeading()
			}
			<-ctx.Done()
			p.setLeader(false)
			for _, callback := range p.callbacks {
				callback.OnStoppedLeading()
			}
			return nil
		})
		p.setLeader(false)
		// in case of error (ex. connection to postgres is dropped) we want to retry the lock with some backoff
		// returning error here would shut down the CP
		if err != nil {
			log.Error(err, "error waiting for lock")
		}

		if util_channels.IsClosed(stop) {
			break
		}
		time.Sleep(backoffTime)
	}
	log.Info("Leader Elector stopped")
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

type KumaPqLockLogger struct {
}

func (k *KumaPqLockLogger) Println(msgParts ...interface{}) {
	stringParts := make([]string, len(msgParts))
	for i, msgPart := range msgParts {
		stringParts[i] = fmt.Sprint(msgPart)
	}
	log.Info(strings.Join(stringParts, " "))
}
