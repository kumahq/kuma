package admin

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	envoy_admin_tls "github.com/kumahq/kuma/pkg/envoy/admin/tls"
)

type EnvoyAdminClient interface {
	PostQuit(ctx context.Context, dataplane *core_mesh.DataplaneResource) error

	Stats(ctx context.Context, proxy core_model.ResourceWithAddress, format v1alpha1.AdminOutputFormat) ([]byte, error)
	Clusters(ctx context.Context, proxy core_model.ResourceWithAddress, format v1alpha1.AdminOutputFormat) ([]byte, error)
	ConfigDump(ctx context.Context, proxy core_model.ResourceWithAddress, includeEds bool) ([]byte, error)
}

type envoyAdminClient struct {
	rm               manager.ResourceManager
	caManagers       ca.Managers
	defaultAdminPort uint32

	caCertPool *x509.CertPool
	clientCert *tls.Certificate
}

func NewEnvoyAdminClient(rm manager.ResourceManager, caManagers ca.Managers, adminPort uint32) EnvoyAdminClient {
	client := &envoyAdminClient{
		rm:               rm,
		caManagers:       caManagers,
		defaultAdminPort: adminPort,
	}
	return client
}

func (a *envoyAdminClient) buildHTTPClient(ctx context.Context) (*http.Client, error) {
	caCertPool, clientCert, err := a.mtlsCerts(ctx)
	if err != nil {
		return nil, err
	}

	c := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 3 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 3 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion:   tls.VersionTLS12,
				RootCAs:      &caCertPool,
				Certificates: []tls.Certificate{clientCert},
			},
		},
		Timeout: 5 * time.Second,
	}
	return c, err
}

func (a *envoyAdminClient) mtlsCerts(ctx context.Context) (x509.CertPool, tls.Certificate, error) {
	if a.caCertPool == nil {
		ca, err := envoy_admin_tls.LoadCA(ctx, a.rm)
		if err != nil {
			return x509.CertPool{}, tls.Certificate{}, errors.Wrap(err, "could not load the CA")
		}
		caCertPool := x509.NewCertPool()
		caCert, err := x509.ParseCertificate(ca.Certificate[0])
		if err != nil {
			return x509.CertPool{}, tls.Certificate{}, errors.Wrap(err, "could not parse CA")
		}
		caCertPool.AddCert(caCert)

		pair, err := envoy_admin_tls.GenerateClientCert(ca)
		if err != nil {
			return x509.CertPool{}, tls.Certificate{}, errors.Wrap(err, "could not generate a client certificate")
		}
		clientCert, err := tls.X509KeyPair(pair.CertPEM, pair.KeyPEM)
		if err != nil {
			return x509.CertPool{}, tls.Certificate{}, errors.Wrap(err, "could not parse the client certificate")
		}

		// cache the certs, so we don't have to load the CA and generate them on every single request.
		// This means that if we want to change Envoy Admin CA, we need to restart all CP instances.
		a.caCertPool = caCertPool
		a.clientCert = &clientCert
	}
	return *a.caCertPool, *a.clientCert, nil
}

const (
	quitquitquit = "quitquitquit"
)

func (a *envoyAdminClient) PostQuit(ctx context.Context, dataplane *core_mesh.DataplaneResource) error {
	httpClient, err := a.buildHTTPClient(ctx)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s/%s", dataplane.AdminAddress(a.defaultAdminPort), quitquitquit)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, http.NoBody)
	if err != nil {
		return err
	}

	// Envoy will not send back any response, so do we not check the response
	response, err := httpClient.Do(request)
	if errors.Is(err, io.EOF) {
		return nil // Envoy may not respond correctly for this request because it already started the shut-down process.
	}
	if err != nil {
		return errors.Wrapf(err, "unable to send POST to %s", quitquitquit)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.Errorf("envoy response [%d %s] [%s]", response.StatusCode, response.Status, response.Body)
	}

	return nil
}

func (a *envoyAdminClient) Stats(ctx context.Context, proxy core_model.ResourceWithAddress, format v1alpha1.AdminOutputFormat) ([]byte, error) {
	query := url.Values{}
	if format == v1alpha1.AdminOutputFormat_JSON {
		query.Add("format", "json")
	}
	return a.executeRequest(ctx, proxy, "stats", query)
}

func (a *envoyAdminClient) Clusters(ctx context.Context, proxy core_model.ResourceWithAddress, format v1alpha1.AdminOutputFormat) ([]byte, error) {
	query := url.Values{}
	if format == v1alpha1.AdminOutputFormat_JSON {
		query.Add("format", "json")
	}
	return a.executeRequest(ctx, proxy, "clusters", query)
}

func (a *envoyAdminClient) ConfigDump(ctx context.Context, proxy core_model.ResourceWithAddress, includeEds bool) ([]byte, error) {
	query := url.Values{}
	if includeEds {
		query.Add("include_eds", "true")
	}
	configDump, err := a.executeRequest(ctx, proxy, "config_dump", query)
	if err != nil {
		return nil, err
	}

	var value []byte
	if value, err = Sanitize(configDump); err != nil {
		return nil, err
	}

	return value, nil
}

func (a *envoyAdminClient) executeRequest(ctx context.Context, proxy core_model.ResourceWithAddress, path string, query url.Values) ([]byte, error) {
	var httpClient *http.Client
	var err error
	u := &url.URL{}

	switch proxy.(type) {
	case *core_mesh.DataplaneResource:
		httpClient, err = a.buildHTTPClient(ctx)
		if err != nil {
			return nil, err
		}
		u.Scheme = "https"
	case *core_mesh.ZoneIngressResource, *core_mesh.ZoneEgressResource:
		httpClient, err = a.buildHTTPClient(ctx)
		if err != nil {
			return nil, err
		}
		u.Scheme = "https"
	default:
		return nil, errors.New("unsupported proxy type")
	}

	if host, _, err := net.SplitHostPort(proxy.AdminAddress(a.defaultAdminPort)); err == nil && host == "127.0.0.1" {
		httpClient = &http.Client{
			Timeout: 5 * time.Second,
		}
		u.Scheme = "http"
	}

	u.Host = proxy.AdminAddress(a.defaultAdminPort)
	u.Path = path
	u.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to send GET to %s", "config_dump")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.Errorf("envoy response [%d %s] [%s]", response.StatusCode, response.Status, response.Body)
	}

	resp, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
