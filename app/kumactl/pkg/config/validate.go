package config

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/api-server/types"
	kumactl_config "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
)

// overridden by tests
var DefaultApiServerTimeout = 5 * time.Second

func ValidateCpCoordinates(cp *kumactl_config.ControlPlane) error {
	req, err := http.NewRequest("GET", cp.Coordinates.ApiServer.Url, nil)
	if err != nil {
		return errors.Wrap(err, "could not construct the request")
	}
	client := http.Client{
		Timeout:   DefaultApiServerTimeout,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "could not connect to the Control Plane API Server")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("Control Plane API Server is not responding")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "could not read body from the Control Plane API Server")
	}
	response := types.IndexResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return errors.Wrap(err, "could not unmarshal body from the Control Plane API Server. Provided address is not valid Kuma Control Plane API Server")
	}
	if response.Tagline != types.TaglineKuma {
		return errors.New("provided address is not valid Kuma Control Plane API Server")
	}
	return nil
}
