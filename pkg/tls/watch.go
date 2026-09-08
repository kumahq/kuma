package tls

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
)

const certWatcherRetryInterval = 10 * time.Second

// Watchers hands out watched key pairs, one per pair of files.
// autoconfigureServersTLS points every control plane server at the general
// certificate unless the server configures its own, so servers reading the same
// files share a watch instead of each running its own over the same two files.
// Sharing also keeps everything that has to agree on what the process serves -
// the certificate a server presents and the CA the bootstrap server hands to a
// data plane proxy - reading the same key pair.
type Watchers struct {
	ctx context.Context
	log logr.Logger

	mtx      sync.Mutex
	watchers map[watchedFiles]*Watcher
}

type watchedFiles struct {
	certFile string
	keyFile  string
}

// NewWatchers returns key pair watchers that stop when ctx is done.
func NewWatchers(ctx context.Context, log logr.Logger) *Watchers {
	return &Watchers{
		ctx:      ctx,
		log:      log,
		watchers: map[watchedFiles]*Watcher{},
	}
}

// Watch loads a key pair from disk and keeps watching the files, so a
// certificate rotated by an external tool (cert-manager, Vault agent, ...) is
// served without restarting the process. Loading fails when the files cannot be
// read or parsed, later failures leave the key pair that is served in place.
// Callers passing the same files get the same watcher.
func (w *Watchers) Watch(certFile string, keyFile string) (*Watcher, error) {
	files := watchedFiles{certFile: certFile, keyFile: keyFile}

	w.mtx.Lock()
	defer w.mtx.Unlock()
	if watcher, ok := w.watchers[files]; ok {
		return watcher, nil
	}

	certWatcher, err := certwatcher.New(certFile, keyFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load TLS certificate")
	}
	log := w.log.WithValues("certFile", certFile, "keyFile", keyFile)
	// certwatcher logs a reload under controller-runtime's own logger and says
	// nothing about the certificate that is served from then on.
	certWatcher.RegisterCallback(func(cert tls.Certificate) {
		if cert.Leaf == nil {
			log.Info("serving TLS certificate")
			return
		}
		log.Info("serving TLS certificate",
			"serialNumber", cert.Leaf.SerialNumber,
			"dnsNames", cert.Leaf.DNSNames,
			"notAfter", cert.Leaf.NotAfter,
		)
	})
	go w.watch(certWatcher, log)

	watcher := &Watcher{certFile: certFile, watcher: certWatcher}
	w.watchers[files] = watcher
	return watcher, nil
}

func (w *Watchers) watch(watcher *certwatcher.CertWatcher, log logr.Logger) {
	// Watching can fail to start, for example when the host runs out of
	// inotify watches. Retrying keeps a transient failure from disabling
	// reloads for the rest of the process lifetime.
	for w.ctx.Err() == nil {
		if err := watcher.Start(w.ctx); err != nil {
			log.Error(err, "could not watch the TLS certificate", "retryIn", certWatcherRetryInterval)
		}
		select {
		case <-w.ctx.Done():
		case <-time.After(certWatcherRetryInterval):
		}
	}
}

// Watcher is a key pair, reloaded when the files it was read from change.
type Watcher struct {
	certFile string
	watcher  *certwatcher.CertWatcher
}

// CertFile is the file the certificate is read from.
func (w *Watcher) CertFile() string {
	return w.certFile
}

// GetCertificate is what tls.Config#GetCertificate expects.
func (w *Watcher) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return w.Certificate()
}

// Certificate returns the last key pair that loaded successfully, which means a
// rotation caught half-written does not break handshakes.
func (w *Watcher) Certificate() (*tls.Certificate, error) {
	cert, err := w.watcher.GetCertificate(nil)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return nil, errors.Errorf("no TLS certificate loaded from %s", w.certFile)
	}
	return cert, nil
}
