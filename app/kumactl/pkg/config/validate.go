package config

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kumahq/kuma/pkg/api-server/types"
	kumactl_config "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/version"
)

func ValidateCpCoordinates(cp *kumactl_config.ControlPlane, timeout time.Duration) error {
	req, err := http.NewRequest("GET", cp.Coordinates.ApiServer.Url, nil)
	if err != nil {
		return fmt.Errorf("could not construct the request: %w", err)
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
		return fmt.Errorf("could not connect to the Control Plane API Server: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("Control Plane API Server is not responding")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read body from the Control Plane API Server: %w", err)
	}
	response := types.IndexResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("could not unmarshal body from the Control Plane API Server. Provided address is not valid Kuma Control Plane API Server: %w", err)
	}
	if response.Tagline != version.Product {
		return fmt.Errorf("this CLI is for %s but the control plane you're connected to is %s. Please use the CLI for your control plane", version.Product, response.Tagline)
	}
	return nil
}
