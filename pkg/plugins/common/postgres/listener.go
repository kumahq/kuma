package postgres

<<<<<<< HEAD
import (
	"fmt"
=======
type Listener interface {
	Notify() chan *Notification
	Error() <-chan error
	Close() error
}
>>>>>>> c0953aefb (fix(kuma-cp): make store changes processing more reliable (#6728))

	"github.com/go-logr/logr"
	"github.com/lib/pq"

	config "github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
)

func NewListener(cfg config.PostgresStoreConfig, log logr.Logger) (*pq.Listener, error) {
	connStr, err := cfg.ConnectionString()
	if err != nil {
		return nil, err
	}
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Error(err, "error occurred", "event", fmt.Sprintf("%v", ev))
			return
		}
		log.Info("event happened", "event", ev)
	}
	listener := pq.NewListener(connStr, cfg.MinReconnectInterval.Duration, cfg.MaxReconnectInterval.Duration, reportProblem)
	if err := listener.Listen("resource_events"); err != nil {
		return nil, err
	}
	return listener, nil
}
