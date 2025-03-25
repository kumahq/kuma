package framework

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"

	"github.com/kumahq/kuma/pkg/config/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/test/framework/kumactl"
	"github.com/kumahq/kuma/test/framework/universal"
)

var _ ControlPlane = &VmUniversalControlPlane{}

type VmUniversalControlPlane struct {
	t            testing.TestingT
	mode         core.CpMode
	name         string
	kumactl      *kumactl.KumactlOptions
	verbose      bool
	cpNetworking *universal.Networking
	setupKumactl bool
}

func NewVmUniversalControlPlane(
	t testing.TestingT,
	mode core.CpMode,
	clusterName string,
	verbose bool,
	networking *universal.Networking,
	apiHeaders []string,
	setupKumactl bool,
) (*VmUniversalControlPlane, error) {
	name := clusterName + "-" + mode
	kumactl := NewKumactlOptionsE2E(t, name, verbose)
	ucp := &VmUniversalControlPlane{
		t:            t,
		mode:         mode,
		name:         name,
		kumactl:      kumactl,
		verbose:      verbose,
		cpNetworking: networking,
		setupKumactl: setupKumactl,
	}
	token, err := ucp.retrieveAdminToken()
	if err != nil {
		return nil, err
	}

	if err := kumactl.KumactlConfigControlPlanesAdd(clusterName, ucp.GetAPIServerAddress(), token, apiHeaders); err != nil {
		return nil, err
	}
	return ucp, nil
}

func (c *VmUniversalControlPlane) Networking() *universal.Networking {
	return c.cpNetworking
}

func (c *VmUniversalControlPlane) GetName() string {
	return c.name
}

func (c *VmUniversalControlPlane) GetKDSInsecureServerAddress() string {
	return c.getKDSServerAddress(false)
}

func (c *VmUniversalControlPlane) GetKDSServerAddress() string {
	return c.getKDSServerAddress(true)
}

func (c *VmUniversalControlPlane) GetXDSServerAddress() string {
	return net.JoinHostPort(c.cpNetworking.IP, "5678")
}

func (c *VmUniversalControlPlane) getKDSServerAddress(secure bool) string {
	protocol := "grpcs"
	if !secure {
		protocol = "grpc"
	}

	return protocol + "://" + net.JoinHostPort(c.cpNetworking.IP, "5685")
}

func (c *VmUniversalControlPlane) GetGlobalStatusAPI() string {
	panic("not implemented")
}

func (c *VmUniversalControlPlane) GetAPIServerAddress() string {
	return "http://localhost:" + c.cpNetworking.ApiServerPort
}

func (c *VmUniversalControlPlane) GetMetrics() (string, error) {
	stdout, stderr, err := c.Exec(
		"curl", "--no-progress-meter",
		"--fail", "--show-error",
		"http://localhost:5680/metrics",
	)
	if err != nil {
		return "", err
	}
	if stderr != "" {
		return "", fmt.Errorf("got on stderr: %q", stderr)
	}
	return stdout, nil
}

func (c *VmUniversalControlPlane) GetMonitoringAssignment(clientId string) (string, error) {
	panic("not implemented")
}

func (c *VmUniversalControlPlane) generateToken(
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
			stdout, stderr, err := c.Exec(
				"curl", "-s",
				"--fail", "--show-error",
				"-H", "\"Content-Type: application/json\"",
				"--data", data,
				"http://localhost:5681/tokens"+tokenPath,
			)
			if err != nil {
				return "", err
			}
			if stderr != "" {
				return "", fmt.Errorf("got stderr %q", stderr)
			}
			return stdout, nil
		},
	)
}

func (c *VmUniversalControlPlane) retrieveAdminToken() (string, error) {
	if !c.setupKumactl {
		return "", nil
	}

	return retry.DoWithRetryE(
		c.t, "fetching user admin token",
		DefaultRetries,
		DefaultTimeout,
		func() (string, error) {
			out, stderr, err := c.Exec("curl", "-s", "--fail", "--show-error", "http://localhost:5681/global-secrets/admin-user-token")
			if err != nil {
				return "", err
			}
			if stderr != "" {
				return "", fmt.Errorf("got content on stderr: %q", stderr)
			}
			return ExtractSecretDataFromResponse(out)
		},
	)
}

func (c *VmUniversalControlPlane) Exec(cmd ...string) (string, string, error) {
	return c.cpNetworking.RunCommand(strings.Join(cmd, " "))
}

func (c *VmUniversalControlPlane) GenerateDpToken(mesh, service string) (string, error) {
	data := fmt.Sprintf(
		`'{"mesh": "%s", "tags": {"kuma.io/service":["%s"]}}'`,
		mesh,
		service,
	)

	return c.generateToken("/dataplane", data)
}

func (c *VmUniversalControlPlane) GenerateZoneIngressToken(zone string) (string, error) {
	data := fmt.Sprintf(`'{"zone": "%s", "scope": ["ingress"]}'`, zone)

	return c.generateToken("/zone", data)
}

func (c *VmUniversalControlPlane) GenerateZoneEgressToken(zone string) (string, error) {
	data := fmt.Sprintf(`'{"zone": "%s", "scope": ["egress"]}'`, zone)

	return c.generateToken("/zone", data)
}

func (c *VmUniversalControlPlane) GenerateZoneToken(
	zone string,
	scope []string,
) (string, error) {
	scopeJson, err := json.Marshal(scope)
	if err != nil {
		return "", err
	}

	data := fmt.Sprintf(`'{"zone": "%s", "scope": %s}'`, zone, scopeJson)

	return c.generateToken("/zone", data)
}

func (c *VmUniversalControlPlane) UpdateObject(
	typeName string,
	objectName string,
	update func(object core_model.Resource) core_model.Resource,
) error {
	return c.kumactl.KumactlUpdateObject(typeName, objectName, update)
}
