package events

import (
	"encoding/json"
	"fmt"
	"time"

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
	out events.Writer
}

func NewListener(cfg postgres.PostgresStoreConfig, out events.Writer) component.Component {
	return &listener{
		cfg: cfg,
		out: out,
	}
}

func (k *listener) Start(stop <-chan struct{}) error {
	listener, err := common_postgres.NewListener(k.cfg, log)
	if err != nil {
		return err
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Error(err, "error closing postgres listener")
		}
	}()

	log.Info("start monitoring")
	for {
		select {
		case n := <-listener.Notify:
			obj := &struct {
				Action string `json:"action"`
				Data   struct {
					Name string `json:"name"`
					Mesh string `json:"mesh"`
					Type string `json:"type"`
				}
			}{}
			if err := json.Unmarshal([]byte(n.Extra), obj); err != nil {
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
				fmt.Println("unknown Action")
				continue
			}
			k.out.Send(op, model.ResourceType(obj.Data.Type), model.ResourceKey{Mesh: obj.Data.Mesh, Name: obj.Data.Name})
		case <-time.After(90 * time.Second):
			log.V(1).Info("received no events for 90 seconds, checking connection")
			go func() {
				if err := listener.Ping(); err != nil {
					log.Error(err, "database ping failed")
				}
			}()
		case <-stop:
			log.Info("stop")
			return nil
		}
	}
}

func (k *listener) NeedLeaderElection() bool {
	return false
}
