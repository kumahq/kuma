package leader

import (
	"fmt"
	"time"

	"cirello.io/pglock"

	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	common_postgres "github.com/kumahq/kuma/pkg/plugins/common/postgres"
	leader_memory "github.com/kumahq/kuma/pkg/plugins/leader/memory"
	leader_postgres "github.com/kumahq/kuma/pkg/plugins/leader/postgres"
)

func NewLeaderElector(b *core_runtime.Builder) (component.LeaderElector, error) {
	switch b.Config().Store.Type {
	case store.PostgresStore:
		db, err := common_postgres.ConnectToDb(*b.Config().Store.Postgres)
		if err != nil {
			return nil, fmt.Errorf("could not connect to postgres: %w", err)
		}
		client, err := pglock.New(db,
			pglock.WithLeaseDuration(5*time.Second),
			pglock.WithHeartbeatFrequency(1*time.Second),
			pglock.WithOwner(b.GetInstanceId()),
			pglock.WithLogger(&leader_postgres.KumaPqLockLogger{}),
		)
		if err != nil {
			return nil, fmt.Errorf("could not create postgres lock client: %w", err)
		}
		elector := leader_postgres.NewPostgresLeaderElector(client)
		return elector, nil
	case store.MemoryStore:
		return leader_memory.NewAlwaysLeaderElector(), nil
	// In case of Kubernetes, Leader Elector is embedded in a Kubernetes ComponentManager
	default:
		return nil, fmt.Errorf("no election leader for storage of type %s", b.Config().Store.Type)
	}
}
