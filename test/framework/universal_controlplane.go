package framework

import (
	"fmt"
	"net"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type UniversalControlPlane struct {
	t       testing.TestingT
	mode    core.CpMode
	name    string
	kumactl *KumactlOptions
	cluster *UniversalCluster
	verbose bool
}

func NewUniversalControlPlane(t testing.TestingT, mode core.CpMode, clusterName string, cluster *UniversalCluster, verbose bool) *UniversalControlPlane {
	name := clusterName + "-" + mode
	kumactl, err := NewKumactlOptions(t, name, verbose)
	if err != nil {
		panic(err)
	}
	return &UniversalControlPlane{
		t:       t,
		mode:    mode,
		name:    name,
		kumactl: kumactl,
		cluster: cluster,
		verbose: verbose,
	}
}

func (c *UniversalControlPlane) GetName() string {
	return c.name
}

func (c *UniversalControlPlane) GetKumaCPLogs() (string, error) {
	return c.cluster.apps[AppModeCP].mainApp.Out(), nil
}

func (c *UniversalControlPlane) GetKDSServerAddress() string {
	return "grpcs://" + net.JoinHostPort(c.cluster.apps[AppModeCP].ip, "5685")
}

func (c *UniversalControlPlane) GetGlobalStatusAPI() string {
	panic("not implemented")
}

func (c *UniversalControlPlane) GetAPIServerAddress() string {
	return "http://localhost:" + c.cluster.apps[AppModeCP].ports["5681"]
}

func (c *UniversalControlPlane) GetMetrics() (string, error) {
	return retry.DoWithRetryE(c.t, "fetching CP metrics", DefaultRetries, DefaultTimeout, func() (string, error) {
		sshApp := NewSshApp(c.verbose, c.cluster.apps[AppModeCP].ports["22"], nil, []string{"curl",
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
			sshApp := NewSshApp(
				c.verbose,
				c.cluster.apps[AppModeCP].ports["22"],
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
