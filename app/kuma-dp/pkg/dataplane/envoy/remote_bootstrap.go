package envoy

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	net_url "net/url"
	"strings"

	envoy_bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
	// import only limited proto definitions from Go Control Plane to save kuma-dp size
	_ "github.com/envoyproxy/go-control-plane/envoy/config/grpc_credential/v2alpha"
	_ "github.com/envoyproxy/go-control-plane/envoy/config/grpc_credential/v3"
	"github.com/golang/protobuf/proto" //lint:ignore SA1019 go-control-plane did not migrate yet
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"

	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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
	DpNotFoundErr = errors.New("Dataplane entity not found. If you are running on Universal please create a Dataplane entity on kuma-cp before starting kuma-dp. If you are running on Kubernetes, please check the kuma-cp logs to determine why the Dataplane entity could not be created by the automatic sidecar injection.")
)

func InvalidRequestErr(msg string) error {
	return errors.Errorf("Invalid request: %s", msg)
}

func IsInvalidRequestErr(err error) bool {
	return strings.HasPrefix(err.Error(), "Invalid request: ")
}

func (b *remoteBootstrap) Generate(url string, cfg kuma_dp.Config, dp *rest_types.Resource) (proto.Message, error) {
	bootstrapUrl, err := net_url.Parse(url)
	if err != nil {
		return nil, err
	}

	if bootstrapUrl.Scheme == "https" {
		if cfg.ControlPlane.CaCert != "" {
			certPool := x509.NewCertPool()
			if ok := certPool.AppendCertsFromPEM([]byte(cfg.ControlPlane.CaCert)); !ok {
				return nil, errors.New("could not add certificate")
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

	backoff, err := retry.NewConstant(cfg.ControlPlane.Retry.Backoff)
	if err != nil {
		return nil, errors.Wrap(err, "could not create retry backoff")
	}
	backoff = retry.WithMaxDuration(cfg.ControlPlane.Retry.MaxDuration, backoff)
	var respBytes []byte
	err = retry.Do(context.Background(), backoff, func(ctx context.Context) error {
		log.Info("trying to fetch bootstrap configuration from the Control Plane")
		respBytes, err = b.requestForBootstrap(bootstrapUrl, cfg, dp)
		if err == nil {
			return nil
		}
		if IsInvalidRequestErr(err) { // there is no point in retrying invalid request
			return err
		}
		switch err {
		case DpNotFoundErr:
			log.Info("Dataplane entity is not yet found in the Control Plane. If you are running on Kubernetes, CP is most likely still in the process of converting Pod to Dataplane. Retrying.", "backoff", cfg.ControlPlane.Retry.Backoff)
		default:
			log.Info("could not fetch bootstrap configuration. Retrying.", "backoff", cfg.ControlPlane.Retry.Backoff, "err", err.Error())
		}
		return retry.RetryableError(err)
	})
	if err != nil {
		return nil, err
	}

	bootstrap := envoy_bootstrap.Bootstrap{}
	if err := util_proto.FromYAML(respBytes, &bootstrap); err != nil {
		return nil, errors.Wrap(err, "could not parse the bootstrap configuration")
	}

	return &bootstrap, nil
}

func (b *remoteBootstrap) requestForBootstrap(url *net_url.URL, cfg kuma_dp.Config, dp *rest_types.Resource) ([]byte, error) {
	url.Path = "/bootstrap"
	var dataplaneResource string
	if dp != nil {
		dpJSON, err := json.Marshal(dp)
		if err != nil {
			return nil, err
		}
		dataplaneResource = string(dpJSON)
	}
	request := types.BootstrapRequest{
		Mesh: cfg.Dataplane.Mesh,
		Name: cfg.Dataplane.Name,
		// if not set in config, the 0 will be sent which will result in providing default admin port
		// that is set in the control plane bootstrap params
		AdminPort:          cfg.Dataplane.AdminPort.Lowest(),
		DataplaneTokenPath: cfg.DataplaneRuntime.TokenPath,
		DataplaneResource:  dataplaneResource,
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
		bodyBytes, err := ioutil.ReadAll(resp.Body)
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
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read the body of the response")
	}
	return respBytes, nil
}
