package envoy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	net_url "net/url"

	"github.com/Kong/kuma/pkg/xds/bootstrap/types"

	kuma_dp "github.com/Kong/kuma/pkg/config/app/kuma-dp"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	envoy_bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

type remoteBootstrap struct {
	client *http.Client
}

func NewRemoteBootstrapGenerator(client *http.Client) BootstrapConfigFactoryFunc {
	rb := remoteBootstrap{client: client}
	return rb.Generate
}

func (b *remoteBootstrap) Generate(url string, cfg kuma_dp.Config) (proto.Message, error) {
	bootstrapUrl, err := net_url.Parse(url)
	if err != nil {
		return nil, err
	}
	bootstrapUrl.Path = "/bootstrap"
	request := types.BootstrapRequest{
		Mesh: cfg.Dataplane.Mesh,
		Name: cfg.Dataplane.Name,
		// if not set in config, the 0 will be sent which will result in providing default admin port
		// that is set in the control plane bootstrap params
		AdminPort:          cfg.Dataplane.AdminPort.Lowest(),
		DataplaneTokenPath: cfg.DataplaneRuntime.TokenPath,
	}
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal request to json")
	}
	resp, err := b.client.Post(bootstrapUrl.String(), "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, errors.Wrap(err, "request to bootstrap server failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return nil, errors.New("status: 404. Did you first apply a Dataplane resource?")
		}
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bootstrap := envoy_bootstrap.Bootstrap{}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read the body of the response")
	}
	if err := util_proto.FromYAML(respBytes, &bootstrap); err != nil {
		return nil, errors.Wrap(err, "could not parse the bootstrap configuration")
	}

	return &bootstrap, nil
}
