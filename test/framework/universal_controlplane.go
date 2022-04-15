package framework

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/test/framework/ssh"
)

type UniversalCPNetworking struct {
	IP            string `json:"ip"` // IP inside a docker network
	ApiServerPort string `json:"apiServerPort"`
	SshPort       string `json:"sshPort"`
}

func (u UniversalCPNetworking) BootstrapAddress() string {
	return "https://" + net.JoinHostPort(u.IP, "5678")
}

type UniversalControlPlane struct {
	t            testing.TestingT
	mode         core.CpMode
	name         string
	kumactl      *KumactlOptions
	verbose      bool
	cpNetworking UniversalCPNetworking
}

func NewUniversalControlPlane(t testing.TestingT, mode core.CpMode, clusterName string, verbose bool, networking UniversalCPNetworking) (*UniversalControlPlane, error) {
	name := clusterName + "-" + mode
	kumactl := NewKumactlOptions(t, name, verbose)
	ucp := &UniversalControlPlane{
		t:            t,
		mode:         mode,
		name:         name,
		kumactl:      kumactl,
		verbose:      verbose,
		cpNetworking: networking,
	}
	token, err := ucp.retrieveAdminToken()
	if err != nil {
		return nil, err
	}

	if err := kumactl.KumactlConfigControlPlanesAdd(clusterName, ucp.GetAPIServerAddress(), token); err != nil {
		return nil, err
	}
	return ucp, nil
}

func (c *UniversalControlPlane) Networking() UniversalCPNetworking {
	return c.cpNetworking
}

func (c *UniversalControlPlane) GetName() string {
	return c.name
}

func (c *UniversalControlPlane) GetKDSServerAddress() string {
	return "grpcs://" + net.JoinHostPort(c.cpNetworking.IP, "5685")
}

func (c *UniversalControlPlane) GetGlobalStatusAPI() string {
	panic("not implemented")
}

func (c *UniversalControlPlane) GetAPIServerAddress() string {
	return "http://localhost:" + c.cpNetworking.ApiServerPort
}

func (c *UniversalControlPlane) GetMetrics() (string, error) {
	return retry.DoWithRetryE(c.t, "fetching CP metrics", DefaultRetries, DefaultTimeout, func() (string, error) {
		sshApp := ssh.NewApp(c.verbose, c.cpNetworking.SshPort, nil, []string{"curl",
			"--fail", "--show-error",
			"http://localhost:5680/metrics"})
		if err := sshApp.Run(); err != nil {
			return "", err
		}
		if sshApp.Err() != "" {
			return "", errors.New(sshApp.Err())
		}
		return sshApp.Out(), nil
	})
}

func (c *UniversalControlPlane) generateToken(
	tokenPath string,
	data string,
) (string, error) {
	description := fmt.Sprintf("generating %s token", tokenPath)

	return retry.DoWithRetryE(
		c.t,
		description,
		DefaultRetries,
		DefaultTimeout,
		func() (string, error) {
			sshApp := ssh.NewApp(
				c.verbose,
				c.cpNetworking.SshPort,
				nil,
				[]string{"curl",
					"--fail", "--show-error",
					"-H", "\"Content-Type: application/json\"",
					"--data", data,
					"http://localhost:5681/tokens" + tokenPath},
			)

			if err := sshApp.Run(); err != nil {
				return "", err
			}

			if sshApp.Err() != "" {
				return "", errors.New(sshApp.Err())
			}

			return sshApp.Out(), nil
		},
	)
}

func (c *UniversalControlPlane) retrieveAdminToken() (string, error) {
	return retry.DoWithRetryE(
		c.t, "fetching user admin token",
		DefaultRetries,
		DefaultTimeout,
		func() (string, error) {
			sshApp := ssh.NewApp(
				c.verbose, c.cpNetworking.SshPort, nil, []string{
					"curl", "--fail", "--show-error",
					"http://localhost:5681/global-secrets/admin-user-token",
				},
			)
			if err := sshApp.Run(); err != nil {
				return "", err
			}
			if sshApp.Err() != "" {
				return "", errors.New(sshApp.Err())
			}
			var secret map[string]string
			if err := json.Unmarshal([]byte(sshApp.Out()), &secret); err != nil {
				return "", err
			}
			data := secret["data"]
			token, err := base64.StdEncoding.DecodeString(data)
			if err != nil {
				return "", err
			}
			return string(token), nil
		},
	)
}

func (c *UniversalControlPlane) GenerateDpToken(mesh, service string) (string, error) {
	dpType := ""
	if service == "ingress" {
		dpType = "ingress"
	}

	data := fmt.Sprintf(
		`'{"mesh": "%s", "type": "%s", "tags": {"kuma.io/service":["%s"]}}'`,
		mesh,
		dpType,
		service,
	)

	return c.generateToken("", data)
}

func (c *UniversalControlPlane) GenerateZoneIngressToken(zone string) (string, error) {
	data := fmt.Sprintf(`'{"zone": "%s"}'`, zone)

	return c.generateToken("/zone-ingress", data)
}

func (c *UniversalControlPlane) GenerateZoneEgressToken(zone string) (string, error) {
	data := fmt.Sprintf(`'{"zone": "%s", "scope": ["egress"]}'`, zone)

	return c.generateToken("/zone", data)
}

func (c *UniversalControlPlane) UpdateObject(
	typeName string,
	objectName string,
	update func(object core_model.Resource) core_model.Resource,
) error {
	return c.kumactl.KumactlUpdateObject(typeName, objectName, update)
}
