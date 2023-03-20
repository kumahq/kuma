package events

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/events"
	common_postgres "github.com/kumahq/kuma/pkg/plugins/common/postgres"
)

var log = core.Log.WithName("postgres-event-listener")

type listener struct {
	cfg postgres.PostgresStoreConfig
	out events.Emitter
}

func NewListener(cfg postgres.PostgresStoreConfig, out events.Emitter) component.Component {
	return &listener{
		cfg: cfg,
		out: out,
	}
}

func (k *listener) Start(stop <-chan struct{}) error {
	connectionString, err := k.cfg.ConnectionString()
	if err != nil {
		return err
	}
	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		return err
	}
	listener, err := common_postgres.NewPgxListener(context.Background(), &log, pool, "resource_events")
	if err != nil {
		return err
	}
	listener.Start()
	defer func() {
		if err := listener.Close(); err != nil {
			log.Error(err, "error closing postgres listener")
		}
	}()

	notifications := listener.Notifications()

	log.Info("start monitoring")
	for {
		select {
		case n := <-notifications:
			if n == nil {
				continue
			}
			obj := &struct {
				Action string `json:"action"`
				Data   struct {
					Name string `json:"name"`
					Mesh string `json:"mesh"`
					Type string `json:"type"`
				}
			}{}
			if err := json.Unmarshal([]byte(n.Payload), obj); err != nil {
				log.Error(err, "unable to unmarshal event from PostgreSQL")
				continue
			}
			var op events.Op
			switch obj.Action {
			case "INSERT":
				op = events.Create
			case "UPDATE":
				op = events.Update
			case "DELETE":
				op = events.Delete
			default:
				log.Error(errors.New("unknown Action"), "failed to parse action", "action", op)
				continue
			}
			k.out.Send(events.ResourceChangedEvent{
				Operation: op,
				Type:      model.ResourceType(obj.Data.Type),
				Key:       model.ResourceKey{Mesh: obj.Data.Mesh, Name: obj.Data.Name},
			})
		case <-stop:
			log.Info("stop")
			return nil
		}
	}
}

func (k *listener) NeedLeaderElection() bool {
	return false
}
