package framework

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"

	"github.com/gruntwork-io/terratest/modules/testing"
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
	panic("not implemented")
}

func (c *UniversalControlPlane) GetKDSServerAddress() string {
	return "grpcs://" + c.cluster.apps[AppModeCP].ip + ":5685"
}

func (c *UniversalControlPlane) GetGlobaStatusAPI() string {
	panic("not implemented")
}

func (c *UniversalControlPlane) GenerateDpToken(mesh, service string) (string, error) {
	dpType := ""
	if service == "ingress" {
		dpType = "ingress"
	}
	return retry.DoWithRetryE(c.t, "generating DP token", DefaultRetries, DefaultTimeout, func() (string, error) {
		sshApp := NewSshApp(c.verbose, c.cluster.apps[AppModeCP].ports["22"], []string{}, []string{"curl",
			"--fail", "--show-error",
			"-H", "\"Content-Type: application/json\"",
			"--data", fmt.Sprintf(`'{"mesh": "%s", "type": "%s", "tags": {"kuma.io/service":["%s"]}}'`, mesh, dpType, service),
			"http://localhost:5681/tokens"})
		if err := sshApp.Run(); err != nil {
			return "", err
		}
		if sshApp.Err() != "" {
			return "", errors.New(sshApp.Err())
		}
		return sshApp.Out(), nil
	})
}
