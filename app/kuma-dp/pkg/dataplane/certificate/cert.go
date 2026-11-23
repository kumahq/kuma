package certificate

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"

	"github.com/kumahq/kuma/v3/pkg/core"
)

var logger = core.Log.WithName("system-certificate-selector")

// BundleFileName is the file in the work directory that holds the system CAs
// collected from a directory trust store
const BundleFileName = "system-ca-bundle.pem"

// Source of CA File Paths: https://golang.org/src/crypto/x509/root_linux.go
var certFiles = []string{
	"/etc/ssl/certs/ca-certificates.crt",                // Debian/Ubuntu/Gentoo etc.
	"/etc/pki/tls/certs/ca-bundle.crt",                  // Fedora/RHEL 6
	"/etc/ssl/ca-bundle.pem",                            // OpenSUSE
	"/etc/pki/tls/cacert.pem",                           // OpenELEC
	"/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem", // CentOS/RHEL 7
	"/etc/ssl/cert.pem",                                 // Alpine Linux
}

var certDirectories = []string{
	"/etc/ssl/certs",     // SLES10/SLES11, https://golang.org/issue/12139
	"/etc/pki/tls/certs", // Fedora/RHEL
}

// GetOsCaFilePath returns a path to the system CA certificates that Envoy can load.
// Like Go, it honors SSL_CERT_FILE and SSL_CERT_DIR. When there is no CA bundle file,
// the certificates of the directory trust store are written to a bundle in workDir.
func GetOsCaFilePath(workDir string) string {
	files := certFiles
	if f := os.Getenv("SSL_CERT_FILE"); f != "" {
		files = []string{f}
	}
	dirs := certDirectories
	if d := os.Getenv("SSL_CERT_DIR"); d != "" {
		dirs = strings.Split(d, ":")
	}
	return findOsCa(files, dirs, workDir)
}

func findOsCa(files, dirs []string, workDir string) string {
	for _, cert := range files {
		if info, err := os.Stat(cert); err == nil && !info.IsDir() { // #nosec G703 -- SSL_CERT_FILE is set by the operator, as for Go
			logger.Info("using OS provided CA certificate", "certificate path", cert)
			return cert
		}
	}

	bundle, count := collectDirCerts(dirs)
	if count == 0 {
		logger.Info("OS provided certificate not found")
		return ""
	}
	if workDir == "" {
		logger.Info("OS provided certificates found in a directory, but there is no work directory to write a bundle to", "directories", dirs)
		return ""
	}
	path := filepath.Join(workDir, BundleFileName)
	if err := os.WriteFile(path, bundle, 0o600); err != nil {
		logger.Error(err, "unable to write a bundle of the OS provided CA certificates", "path", path)
		return ""
	}
	logger.Info("using OS provided CA certificates from a directory", "directories", dirs, "certificates", count, "certificate path", path)
	return path
}

// collectDirCerts returns the PEM encoded certificates found in the files of dirs, without duplicates
func collectDirCerts(dirs []string) ([]byte, int) {
	var bundle []byte
	seen := map[string]struct{}{}
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			data, err := os.ReadFile(filepath.Join(dir, entry.Name())) // #nosec G703 -- SSL_CERT_DIR is set by the operator, as for Go
			if err != nil {
				// directories, broken links and unreadable files are skipped, as Go does
				continue
			}
			for block, rest := pem.Decode(data); block != nil; block, rest = pem.Decode(rest) {
				if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
					continue
				}
				if _, err := x509.ParseCertificate(block.Bytes); err != nil {
					continue
				}
				if _, ok := seen[string(block.Bytes)]; ok {
					continue
				}
				seen[string(block.Bytes)] = struct{}{}
				bundle = append(bundle, pem.EncodeToMemory(block)...)
			}
		}
	}
	return bundle, len(seen)
}
