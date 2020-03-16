package ca

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
        tls1 "crypto/tls"
  
	"github.com/pkg/errors"

	kumactl_config "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/Kong/kuma/pkg/core/ca/provided/rest/types"
	error_types "github.com/Kong/kuma/pkg/core/rest/errors/types"
	"github.com/Kong/kuma/pkg/tls"
	util_http "github.com/Kong/kuma/pkg/util/http"
)

const (
	timeout = 10 * time.Second
)

type ProvidedCaClient interface {
	AddSigningCertificate(mesh string, pair tls.KeyPair) (types.SigningCert, error)
	DeleteSigningCertificate(mesh string, id string) error
	SigningCertificates(mesh string) ([]types.SigningCert, error)
}

type httpProvidedCaClient struct {
	client util_http.Client
}

func NewProvidedCaClient(address string, config *kumactl_config.Context_AdminApiCredentials) (ProvidedCaClient, error) {
	baseURL, err := url.Parse(address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse the server URL")
	}
	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{TLSClientConfig: &tls1.Config{InsecureSkipVerify: true},},
	}
	if baseURL.Scheme == "https" {
		if !config.HasClientCert() {
			return nil, errors.New("certificates has to be configured to use https destination")
		}
		// Since we're not going to pass any secrets to the server, we can skip validating its identity.
		if err := util_http.ConfigureTlsWithoutServerVerification(httpClient, config.ClientCert, config.ClientKey); err != nil {
			return nil, errors.Wrap(err, "could not configure tls for provided ca client")
		}
	}
	client := util_http.ClientWithBaseURL(httpClient, baseURL)
	return &httpProvidedCaClient{
		client: client,
	}, nil
}

var _ ProvidedCaClient = &httpProvidedCaClient{}

func (h *httpProvidedCaClient) AddSigningCertificate(mesh string, pair tls.KeyPair) (types.SigningCert, error) {
	urlCerts := fmt.Sprintf("/meshes/%s/ca/provided/certificates", mesh)
	keyPair := types.KeyPair{
		Key:  string(pair.KeyPEM),
		Cert: string(pair.CertPEM),
	}
	pairBytes, err := json.Marshal(keyPair)
	if err != nil {
		return types.SigningCert{}, err
	}
	req, err := http.NewRequest("POST", urlCerts, bytes.NewReader(pairBytes))
	if err != nil {
		return types.SigningCert{}, err
	}
	req.Header.Add("content-type", "application/json")
	respBytes, err := h.doRequest(req)
	if err != nil {
		return types.SigningCert{}, err
	}
	signingCert := types.SigningCert{}
	if err := json.Unmarshal(respBytes, &signingCert); err != nil {
		return types.SigningCert{}, err
	}
	return signingCert, nil
}

func (h *httpProvidedCaClient) SigningCertificates(mesh string) ([]types.SigningCert, error) {
	urlCerts := fmt.Sprintf("/meshes/%s/ca/provided/certificates", mesh)
	req, err := http.NewRequest("GET", urlCerts, nil)
	if err != nil {
		return nil, err
	}
	body, err := h.doRequest(req)
	if err != nil {
		return nil, err
	}
	var certs []types.SigningCert
	if err := json.Unmarshal(body, &certs); err != nil {
		return nil, err
	}
	return certs, nil
}

func (h *httpProvidedCaClient) DeleteSigningCertificate(mesh string, id string) error {
	urlCerts := fmt.Sprintf("/meshes/%s/ca/provided/certificates/%s", mesh, id)
	req, err := http.NewRequest("DELETE", urlCerts, nil)
	if err != nil {
		return err
	}
	_, err = h.doRequest(req)
	return err
}

func (h *httpProvidedCaClient) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Accept", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 >= 4 {
		kumaErr := error_types.Error{}
		if err := json.Unmarshal(b, &kumaErr); err == nil {
			if kumaErr.Title != "" && kumaErr.Details != "" {
				return nil, &kumaErr
			}
		}
		return nil, errors.Errorf("(%d): %s", resp.StatusCode, string(b))
	}
	return b, nil
}
