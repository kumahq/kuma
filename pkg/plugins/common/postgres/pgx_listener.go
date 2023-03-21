package postgres

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
)

// PgxListener will listen for NOTIFY commands on a channel.
type PgxListener struct {
	notificationsCh chan *Notification

	logger logr.Logger

	channel string
	conn    *pgxpool.Conn
	ctx     context.Context
	db      *pgxpool.Pool

	stopFn    func()
	stoppedCh chan struct{}

	mx      sync.Mutex
	running bool
}

var _ Listener = (*PgxListener)(nil)

// NewPgxListener will create and initialize a PgxListener which will automatically reconnect and listen to the provided channels.
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
		ctx:             ctx,
		channel:         "resource_events",
		db:              db,
	}

	err = l.connect(ctx)
	if err != nil {
		return nil, err
	}

	l.Start()

	return l, nil
}

func (l *PgxListener) Start() {
	l.logger.Info("started")

	l.mx.Lock()
	defer l.mx.Unlock()

	if l.running {
		return
	}

	ctx, cancel := context.WithCancel(l.ctx)
	l.stopFn = cancel
	l.stoppedCh = make(chan struct{}, 1)

	go l.run(ctx)

	l.running = true
}

func (l *PgxListener) Close() error {
	l.stop()
	close(l.notificationsCh)
	return nil
}

func (l *PgxListener) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(l.stoppedCh)
			return
		default:
		}
		err := l.handleNotifications(ctx)
		if errors.Is(err, context.Canceled) {
			err = nil
		}
		if err != nil {
			l.logger.Error(err, "error handling notification")
		}
	}
}

// stop will end all current connections and stop reconnecting.
func (l *PgxListener) stop() {
	l.mx.Lock()
	defer l.mx.Unlock()

	if !l.running {
		return
	}

	l.stopFn()
	<-l.stoppedCh

	l.running = false
	l.logger.Info("stopped")
}

func (l *PgxListener) handleNotifications(ctx context.Context) error {
	defer l.disconnect()
	t := time.NewTicker(3 * time.Second)
	defer t.Stop()
	for {
		err := l.connect(ctx)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err != nil {
			l.logger.Error(err, "error connecting")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-t.C:
				continue
			}
		}
		break
	}

	notification, err := l.conn.Conn().WaitForNotification(ctx)
	if err != nil {
		return err
	}

	l.logger.Info("event happened", "event", notification)
	l.notificationsCh <- toBareNotification(notification)

	return nil
}

func toBareNotification(notification *pgconn.Notification) *Notification {
	return &Notification{
		Payload: notification.Payload,
	}
}

func (l *PgxListener) Notify() chan *Notification { return l.notificationsCh }

func (l *PgxListener) disconnect() {
	if l.conn == nil {
		return
	}
	l.conn.Release()
	l.conn = nil
}

func (l *PgxListener) connect(ctx context.Context) error {
	if l.conn != nil {
		return nil
	}
	conn, err := l.db.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "error getting connection")
	}

	l.conn = conn

	_, err = conn.Exec(ctx, "listen "+l.channel)
	if err != nil {
		l.disconnect()
		return err
	}

	return nil
}
