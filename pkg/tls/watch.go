package tls

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
)

const certWatcherRetryInterval = 10 * time.Second

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
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		<-stop
	}()
	go func() {
		// Watching can fail to start, for example when the host runs out of
		// inotify watches. Retrying keeps a transient failure from disabling
		// reloads for the rest of the process lifetime.
		for ctx.Err() == nil {
			if err := watcher.Start(ctx); err != nil {
				log.Error(err, "could not watch the TLS certificate", "retryIn", certWatcherRetryInterval)
			}
			select {
			case <-ctx.Done():
			case <-time.After(certWatcherRetryInterval):
			}
		}
	}()
	return watcher.GetCertificate, nil
}
