package http

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

func ConfigureTls(httpClient *http.Client, serverCert string, clientCert string, clientKey string) error {
	certBytes, err := ioutil.ReadFile(serverCert)
	if err != nil {
		return errors.Wrap(err, "could not read server cert")
	}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(certBytes); !ok {
		return errors.New("could not add certificate")
	}

	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		return errors.Wrap(err, "could not create key pair from client cert and client key")
	}

	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      certPool,
			Certificates: []tls.Certificate{cert},
		},
	}
	return nil
}

func ConfigureTlsWithoutServerVerification(httpClient *http.Client, clientCert string, clientKey string) error {
	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		return errors.Wrap(err, "could not create key pair from client cert and client key")
	}

	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		},
	}
	return nil
}
