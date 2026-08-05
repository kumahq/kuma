package tls

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"

	"github.com/kumahq/kuma/v3/pkg/core/runtime/component"
)

const (
	certWatcherBackoffBaseTime = 5 * time.Second
	certWatcherBackoffMaxTime  = time.Minute
)

// GetCertificateFunc is what tls.Config#GetCertificate expects.
type GetCertificateFunc func(*tls.ClientHelloInfo) (*tls.Certificate, error)

// WatchKeyPair loads a key pair from disk and keeps watching the files, so a
// certificate rotated by an external tool (cert-manager, Vault agent, ...) is
// served without restarting the process. The returned function is meant to be
// plugged into tls.Config#GetCertificate, it serves the last key pair that was
// loaded successfully, which means a rotation caught half-written does not
// break handshakes. Watching stops when stop is closed.
func WatchKeyPair(certFile string, keyFile string, stop <-chan struct{}, log logr.Logger) (GetCertificateFunc, error) {
	watcher, err := certwatcher.New(certFile, keyFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load TLS certificate")
	}
	log = log.WithName("cert-watcher").WithValues("certFile", certFile, "keyFile", keyFile)
	// Watching can fail to start, for example when the host runs out of
	// inotify watches. Retrying keeps a transient failure from silently
	// disabling reloads for the rest of the process lifetime.
	resilient := component.NewResilientComponent(
		log,
		&certWatcherComponent{watcher: watcher},
		certWatcherBackoffBaseTime,
		certWatcherBackoffMaxTime,
	)
	go func() {
		if err := resilient.Start(stop); err != nil {
			log.Error(err, "certificate watcher terminated with an error, the certificate will not be reloaded")
		}
	}()
	return watcher.GetCertificate, nil
}

// certWatcherComponent adapts the context taken by certwatcher to the stop
// channel a component is started with.
type certWatcherComponent struct {
	watcher *certwatcher.CertWatcher
}

var _ component.Component = &certWatcherComponent{}

func (c *certWatcherComponent) Start(stop <-chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-stop:
			cancel()
		case <-ctx.Done():
		}
	}()
	return c.watcher.Start(ctx)
}

func (c *certWatcherComponent) NeedLeaderElection() bool {
	return false
}
