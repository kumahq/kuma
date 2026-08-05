package tls

import (
	"crypto/tls"
	"os"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

// KeyPairReloader serves a certificate loaded from disk and picks up a new
// one when the files change, so certificates rotated by an external tool
// (cert-manager, Vault agent, ...) are used without restarting the process.
type KeyPairReloader struct {
	certFile string
	keyFile  string
	log      logr.Logger

	mu    sync.Mutex
	cert  *tls.Certificate
	stamp keyPairStamp
}

// keyPairStamp identifies a version of the key pair on disk. Content hashing
// would be more precise, but modification time and size change on every
// rotation, including the symlink swap Kubernetes does on mounted secrets.
type keyPairStamp struct {
	certModTime time.Time
	certSize    int64
	keyModTime  time.Time
	keySize     int64
}

// NewKeyPairReloader loads the key pair and returns a reloader for it. It
// fails when the initial load fails, so a misconfigured path is still
// reported at startup.
func NewKeyPairReloader(certFile string, keyFile string, log logr.Logger) (*KeyPairReloader, error) {
	r := &KeyPairReloader{
		certFile: certFile,
		keyFile:  keyFile,
		log:      log,
	}
	if err := r.reloadIfChanged(); err != nil {
		return nil, errors.Wrap(err, "failed to load TLS certificate")
	}
	return r, nil
}

// GetCertificate is meant to be plugged into tls.Config#GetCertificate. It
// reloads the key pair when it changed on disk since the last handshake.
func (r *KeyPairReloader) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.reloadIfChanged(); err != nil {
		// A rotation that is half-written on disk must not break handshakes,
		// the previous certificate is still valid for a while. The next
		// handshake retries because the stamp was not advanced.
		r.log.Error(err, "could not reload TLS certificate, serving the previously loaded one",
			"certFile", r.certFile, "keyFile", r.keyFile)
	}
	return r.cert, nil
}

// reloadIfChanged reloads the key pair unless the files are unchanged.
// Callers other than NewKeyPairReloader must hold r.mu.
func (r *KeyPairReloader) reloadIfChanged() error {
	stamp, err := statKeyPair(r.certFile, r.keyFile)
	if err != nil {
		return err
	}
	if r.cert != nil && stamp == r.stamp {
		return nil
	}
	cert, err := tls.LoadX509KeyPair(r.certFile, r.keyFile)
	if err != nil {
		return err
	}
	if r.cert != nil {
		r.log.Info("reloaded TLS certificate", "certFile", r.certFile, "keyFile", r.keyFile)
	}
	r.cert = &cert
	r.stamp = stamp
	return nil
}

func statKeyPair(certFile string, keyFile string) (keyPairStamp, error) {
	certInfo, err := os.Stat(certFile)
	if err != nil {
		return keyPairStamp{}, err
	}
	keyInfo, err := os.Stat(keyFile)
	if err != nil {
		return keyPairStamp{}, err
	}
	return keyPairStamp{
		certModTime: certInfo.ModTime(),
		certSize:    certInfo.Size(),
		keyModTime:  keyInfo.ModTime(),
		keySize:     keyInfo.Size(),
	}, nil
}
