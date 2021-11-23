package http

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

func ConfigureMTLS(httpClient *http.Client, caCert string, clientCert string, clientKey string) error {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}

	if caCert == "" {
		transport.TLSClientConfig.InsecureSkipVerify = true
	} else {
		certBytes, err := os.ReadFile(caCert)
		if err != nil {
			return errors.Wrap(err, "could not read CA cert")
		}
		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(certBytes); !ok {
			return errors.New("could not add certificate")
		}
		transport.TLSClientConfig.RootCAs = certPool
	}

	if clientKey != "" && clientCert != "" {
		cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
		if err != nil {
			return errors.Wrap(err, "could not create key pair from client cert and client key")
		}
		transport.TLSClientConfig.Certificates = []tls.Certificate{cert}
	}

	httpClient.Transport = transport
	return nil
}
