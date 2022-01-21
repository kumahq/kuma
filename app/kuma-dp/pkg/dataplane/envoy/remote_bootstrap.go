package envoy

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"
	net_url "net/url"
	"os"
	"strings"

	envoy_bootstrap_v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"

	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	kuma_version "github.com/kumahq/kuma/pkg/version"
	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
)

type remoteBootstrap struct {
	client *http.Client
}

func NewRemoteBootstrapGenerator(client *http.Client) BootstrapConfigFactoryFunc {
	rb := remoteBootstrap{client: client}
	return rb.Generate
}

var (
	log           = core.Log.WithName("dataplane")
	DpNotFoundErr = errors.New("Dataplane entity not found. If you are running on Universal please create a Dataplane entity on kuma-cp before starting kuma-dp or pass it to kuma-dp run --dataplane-file=/file. If you are running on Kubernetes, please check the kuma-cp logs to determine why the Dataplane entity could not be created by the automatic sidecar injection.")
)

func InvalidRequestErr(msg string) error {
	return errors.Errorf("Invalid request: %s", msg)
}

func IsInvalidRequestErr(err error) bool {
	return strings.HasPrefix(err.Error(), "Invalid request: ")
}

func (b *remoteBootstrap) Generate(url string, cfg kuma_dp.Config, params BootstrapParams) (*envoy_bootstrap_v3.Bootstrap, []byte, error) {
	bootstrapUrl, err := net_url.Parse(url)
	if err != nil {
		return nil, nil, err
	}

	if bootstrapUrl.Scheme == "https" {
		if cfg.ControlPlane.CaCert != "" {
			certPool := x509.NewCertPool()
			if ok := certPool.AppendCertsFromPEM([]byte(cfg.ControlPlane.CaCert)); !ok {
				return nil, nil, errors.New("could not add certificate")
			}
			b.client.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: certPool,
				},
			}
		} else {
			log.Info(`[WARNING] The data plane proxy cannot verify the identity of the control plane because you are not setting the "--ca-cert-file" argument or setting the KUMA_CONTROL_PLANE_CA_CERT environment variable.`)
		}
	}

	backoff := retry.WithMaxDuration(cfg.ControlPlane.Retry.MaxDuration, retry.NewConstant(cfg.ControlPlane.Retry.Backoff))
	var respBytes []byte
	err = retry.Do(context.Background(), backoff, func(ctx context.Context) error {
		log.Info("trying to fetch bootstrap configuration from the Control Plane")
		respBytes, err = b.requestForBootstrap(bootstrapUrl, cfg, params)
		if err == nil {
			return nil
		}
		if IsInvalidRequestErr(err) { // there is no point in retrying invalid request
			return err
		}

		switch err {
		case DpNotFoundErr:
			log.Info("Dataplane entity is not yet found in the Control Plane. If you are running on Kubernetes, CP is most likely still in the process of converting Pod to Dataplane. If it takes too long, check kuma-cp logs. Retrying.", "backoff", cfg.ControlPlane.Retry.Backoff)
		default:
			log.Info("could not fetch bootstrap configuration. Retrying.", "backoff", cfg.ControlPlane.Retry.Backoff, "err", err.Error())
		}
		return retry.RetryableError(err)
	})
	if err != nil {
		return nil, nil, err
	}

	bootstrap := &envoy_bootstrap_v3.Bootstrap{}
	if err := util_proto.FromYAML(respBytes, bootstrap); err != nil {
		return nil, nil, err
	}
	return bootstrap, respBytes, nil
}

func (b *remoteBootstrap) requestForBootstrap(url *net_url.URL, cfg kuma_dp.Config, params BootstrapParams) ([]byte, error) {
	url.Path = "/bootstrap"
	var dataplaneResource string
	if params.Dataplane != nil {
		dpJSON, err := json.Marshal(params.Dataplane)
		if err != nil {
			return nil, err
		}
		dataplaneResource = string(dpJSON)
	}
	token := ""
	if cfg.DataplaneRuntime.TokenPath != "" {
		tokenData, err := os.ReadFile(cfg.DataplaneRuntime.TokenPath)
		if err != nil {
			return nil, err
		}
		token = string(tokenData)
	}
	if cfg.DataplaneRuntime.Token != "" {
		token = cfg.DataplaneRuntime.Token
	}
	request := types.BootstrapRequest{
		Mesh:      cfg.Dataplane.Mesh,
		Name:      cfg.Dataplane.Name,
		ProxyType: cfg.Dataplane.ProxyType,
		// if not set in config, the 0 will be sent which will result in providing default admin port
		// that is set in the control plane bootstrap params
		AdminPort:         cfg.Dataplane.AdminPort.Lowest(),
		DataplaneToken:    token,
		DataplaneResource: dataplaneResource,
		BootstrapVersion:  types.BootstrapV3, // set BootstrapVersion to be compatible with old Kuma CPs
		CaCert:            cfg.ControlPlane.CaCert,
		Version: types.Version{
			KumaDp: types.KumaDpVersion{
				Version:   kuma_version.Build.Version,
				GitTag:    kuma_version.Build.GitTag,
				GitCommit: kuma_version.Build.GitCommit,
				BuildDate: kuma_version.Build.BuildDate,
			},
			Envoy: types.EnvoyVersion{
				Version: params.EnvoyVersion.Version,
				Build:   params.EnvoyVersion.Build,
			},
		},
		DynamicMetadata: params.DynamicMetadata,
		DNSPort:         params.DNSPort,
		EmptyDNSPort:    params.EmptyDNSPort,
	}
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal request to json")
	}
	resp, err := b.client.Post(url.String(), "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, errors.Wrap(err, "request to bootstrap server failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to read the response with status code: %d. Make sure you are using https URL", resp.StatusCode)
		}
		if resp.StatusCode == http.StatusNotFound && len(bodyBytes) == 0 {
			return nil, DpNotFoundErr
		}
		if resp.StatusCode == http.StatusNotFound && string(bodyBytes) == "404: Page Not Found" { // response body of Go HTTP Server when hit for invalid endpoint
			return nil, errors.New("There is no /bootstrap endpoint for provided CP address. Double check if the address passed to the CP has a DP Server port (5678 by default), not HTTP API (5681 by default)")
		}
		if resp.StatusCode/100 == 4 {
			return nil, InvalidRequestErr(string(bodyBytes))
		}
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read the body of the response")
	}
	return respBytes, nil
}
