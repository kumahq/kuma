package certificate

import (
	"os"

	"github.com/kumahq/kuma/pkg/core"
)

var logger = core.Log.WithName("system-certificate-selector")

func GetOsCaFilePath() string {
	// Source of CA File Paths: https://golang.org/src/crypto/x509/root_linux.go
	certFiles := []string{
		"/etc/ssl/certs/ca-certificates.crt",                // Debian/Ubuntu/Gentoo etc.
		"/etc/pki/tls/certs/ca-bundle.crt",                  // Fedora/RHEL 6
		"/etc/ssl/ca-bundle.pem",                            // OpenSUSE
		"/etc/pki/tls/cacert.pem",                           // OpenELEC
		"/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem", // CentOS/RHEL 7
		"/etc/ssl/cert.pem",                                 // Alpine Linux
	}

	for _, cert := range certFiles {
		if _, err := os.Stat(cert); err == nil {
			logger.Info("using OS provided CA certificate", "certificate path", cert)
			return cert
		}
	}
	logger.Info("OS provided certificate not found")
	return ""
}
