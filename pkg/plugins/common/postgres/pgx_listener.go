package postgres

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
)

// PgxListener will listen for NOTIFY commands on a channel.
type PgxListener struct {
	notificationsCh chan *Notification
	err             error

	logger logr.Logger

	db *pgxpool.Pool

	stopFn func()
}

func (l *PgxListener) Error() error {
	return l.err
}

var _ Listener = (*PgxListener)(nil)

// NewPgxListener will create and initialize a PgxListener which will automatically connect and listen to the provided channel.
func NewPgxListener(config postgres.PostgresStoreConfig, logger logr.Logger) (Listener, error) {
	ctx := context.Background()
	connectionString, err := config.ConnectionString()
	if err != nil {
		return nil, err
	}
	db, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		return nil, err
	}
	l := &PgxListener{
		notificationsCh: make(chan *Notification, 32),
		logger:          logger,
		db:              db,
	}
	l.start(ctx)

	return l, nil
}

func (l *PgxListener) start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	l.stopFn = cancel

	go l.run(ctx)
	l.logger.Info("started")
}

func (l *PgxListener) Close() error {
	l.stopFn()
	close(l.notificationsCh)
	l.logger.Info("stopped")
	return nil
}

func (l *PgxListener) run(ctx context.Context) {
	err := l.handleNotifications(ctx)
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	if err != nil {
		l.err = err
		close(l.notificationsCh)
		return
	}
}

func (l *PgxListener) handleNotifications(ctx context.Context) error {
	conn, err := l.db.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "error getting connection")
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "listen "+channelName)
	if err != nil {
		return err
	}

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return err
		}

		l.logger.V(1).Info("event happened", "event", notification)
		l.notificationsCh <- toBareNotification(notification)
	}
}

func toBareNotification(notification *pgconn.Notification) *Notification {
	return &Notification{
		Payload: notification.Payload,
	}
}

func (l *PgxListener) Notify() chan *Notification { return l.notificationsCh }
