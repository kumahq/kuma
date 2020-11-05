package postgres

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/lib/pq"

	config "github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
)

func NewListener(cfg config.PostgresStoreConfig, log logr.Logger) (*pq.Listener, error) {
	connStr, err := connectionString(cfg)
	if err != nil {
		return nil, err
	}
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Error(err, "error occurred")
		}
	}
	listener := pq.NewListener(connStr, 10*time.Second, time.Minute, reportProblem)
	if err := listener.Listen("events"); err != nil {
		return nil, err
	}
	return listener, nil
}
