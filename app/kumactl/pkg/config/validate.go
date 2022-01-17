package config

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/api-server/types"
	kumactl_config "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/version"
)

func ValidateCpCoordinates(cp *kumactl_config.ControlPlane, timeout time.Duration) error {
	req, err := http.NewRequest("GET", cp.Coordinates.ApiServer.Url, nil)
	if err != nil {
		return errors.Wrap(err, "could not construct the request")
	}
	client := http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	for _, h := range cp.Coordinates.ApiServer.Headers {
		req.Header.Add(h.Key, h.Value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "could not connect to the Control Plane API Server")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("Control Plane API Server is not responding")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "could not read body from the Control Plane API Server")
	}
	response := types.IndexResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return errors.Wrap(err, "could not unmarshal body from the Control Plane API Server. Provided address is not valid Kuma Control Plane API Server")
	}
	if response.Tagline != version.Product {
		return errors.New("provided address is not valid Kuma Control Plane API Server")
	}
	return nil
}
