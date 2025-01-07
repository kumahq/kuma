package client

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/url"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type TLSConfig struct {
	CaCert     x509.Certificate
	ClientCert tls.Certificate
}

type Conn interface {
	grpc.ClientConnInterface
	io.Closer
	GetState() connectivity.State
}

func New(serverURL string, tlsCfg *TLSConfig) (Conn, error) {
	url, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	var dialOpts []grpc.DialOption
	switch url.Scheme {
	case "grpc": // not used in production
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	case "grpcs":
		tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
		if tlsCfg != nil {
			cp := x509.NewCertPool()
			cp.AddCert(&tlsCfg.CaCert)
			tlsConfig.RootCAs = cp
			tlsConfig.Certificates = []tls.Certificate{tlsCfg.ClientCert}
		} else {
			tlsConfig.InsecureSkipVerify = true
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	default:
		return nil, errors.Errorf("unsupported scheme %q. Use one of %s", url.Scheme, []string{"grpc", "grpcs"})
	}
	return grpc.NewClient(url.Host, dialOpts...)
}
