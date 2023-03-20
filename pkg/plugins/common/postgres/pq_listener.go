package postgres

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/lib/pq"

	config "github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
)

type pqListener struct {
	listener *pq.Listener
	notifications chan *Notification
}

var _ Listener = (*pqListener)(nil)

func (p *pqListener) Notify() chan *Notification {
	return p.notifications
}

func (p *pqListener) Close() error {
	defer close(p.notifications)
	return p.listener.Close()
}

func NewListener(cfg config.PostgresStoreConfig, log logr.Logger) (Listener, error) {
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

	return toListener(listener), nil
}

func toListener(listener *pq.Listener) Listener {
	pqNotificationCh := listener.NotificationChannel()
	notificationCh := make(chan *Notification)

	go func() {
		for {
			pqNotification, more := <- pqNotificationCh
			if more {
				notification := toNotification(pqNotification)
				notificationCh <- notification
			}
		}
	}()

	return &pqListener{
		listener:      listener,
		notifications: notificationCh,
	}
}

func toNotification(pqNotification *pq.Notification) *Notification {
	return &Notification{
		Payload: pqNotification.Extra,
	}
}
