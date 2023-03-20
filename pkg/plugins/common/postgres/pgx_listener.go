package postgres

import (
	"context"
	"sync"
	"time"

	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// PgxListener will listen for NOTIFY commands on a channel.
type PgxListener struct {
	errCh           chan error
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
		errCh:           make(chan error),
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

// Start will enable reconnections and messages.
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

func (l *PgxListener) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			l.stoppedCh <- struct{}{}
			return
		default:
		}
		err := l.handleNotifications(ctx)
		if errors.Is(err, context.Canceled) {
			err = nil
		}
		if err != nil {
			l.errCh <- err
		}
	}
}

// Stop will end all current connections and stop reconnecting.
func (l *PgxListener) Stop() {
	l.logger.Info("stopped")
	l.mx.Lock()
	defer l.mx.Unlock()

	if !l.running {
		return
	}

	l.stopFn()
	<-l.stoppedCh

	l.running = false
}

// Close performs a shutdown with a background context.
func (l *PgxListener) Close() error {
	return l.Shutdown(context.Background())
}

// Shutdown will shut down the listener and returns after all connections have been completed.
// It is not necessary to call Stop() before Close().
func (l *PgxListener) Shutdown(context.Context) error {
	l.Stop()
	close(l.notificationsCh)
	close(l.errCh)
	return nil
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
			l.errCh <- err
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

	select {
	// Writing to channel
	case l.notificationsCh <- toBareNotification(notification):
	default:
	}

	return nil
}

func toBareNotification(notification *pgconn.Notification) *Notification {
	return &Notification{
		Payload: notification.Payload,
	}
}

// Errors will return a channel that will be fed errors from this listener.
func (l *PgxListener) Errors() <-chan error { return l.errCh }

// Notify returns the notification channel for this listener.
// Nil values will not be returned until the listener is closed.
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
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	conn, err := l.db.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "get connection")
	}

	l.conn = conn

	select {
	case <-ctx.Done():
		l.disconnect()
		return ctx.Err()
	default:
	}

	_, err = conn.Exec(ctx, "listen "+l.channel)
	if err != nil {
		l.disconnect()
		return err
	}

	return nil
}
