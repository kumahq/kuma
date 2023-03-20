package postgres

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// Listener will listen for NOTIFY commands on a set of channels.
type Listener struct {
	notifCh chan *pgconn.Notification

	logger *logr.Logger

	ctx      context.Context
	db       *pgxpool.Pool
	channels []string

	stopFn    func()
	stoppedCh chan struct{}
	errCh     chan error

	mx      sync.Mutex
	conn    *pgxpool.Conn
	running bool
}

// NewPgxListener will create and initialize a Listener which will automatically reconnect and listen to the provided channels.
func NewPgxListener(ctx context.Context, logger *logr.Logger, db *pgxpool.Pool, channels ...string) (*Listener, error) {
	l := &Listener{
		notifCh:  make(chan *pgconn.Notification, 32),
		ctx:      ctx,
		channels: channels,
		db:       db,
		errCh:    make(chan error),
		logger:   logger,
	}

	err := l.connect(ctx)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// Start will enable reconnections and messages.
func (l *Listener) Start() {
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

func (l *Listener) run(ctx context.Context) {
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
func (l *Listener) Stop() {
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
func (l *Listener) Close() error {
	return l.Shutdown(context.Background())
}

// Shutdown will shut down the listener and returns after all connections have been completed.
// It is not necessary to call Stop() before Close().
func (l *Listener) Shutdown(context.Context) error {
	l.Stop()
	close(l.notifCh)
	close(l.errCh)
	return nil
}

func (l *Listener) handleNotifications(ctx context.Context) error {
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
	case l.notifCh <- notification:
	default:
	}

	return nil
}

// Errors will return a channel that will be fed errors from this listener.
func (l *Listener) Errors() <-chan error { return l.errCh }

// Notifications returns the notification channel for this listener.
// Nil values will not be returned until the listener is closed.
func (l *Listener) Notifications() <-chan *pgconn.Notification { return l.notifCh }

func (l *Listener) disconnect() {
	if l.conn == nil {
		return
	}
	l.conn.Release()
	l.conn = nil
}

func (l *Listener) connect(ctx context.Context) error {
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

	for _, name := range l.channels {
		select {
		case <-ctx.Done():
			l.disconnect()
			return ctx.Err()
		default:
		}

		_, err = conn.Exec(ctx, "listen "+quoteID(name))
		if err != nil {
			l.disconnect()
			return err
		}
	}

	return nil
}

func quoteID(parts ...string) string {
	return pgx.Identifier(parts).Sanitize()
}
