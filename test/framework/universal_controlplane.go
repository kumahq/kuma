package framework

import (
	"github.com/gruntwork-io/terratest/modules/testing"

	"github.com/Kong/kuma/pkg/config/mode"
)

type UniversalControlPlane struct {
	t       testing.TestingT
	mode    mode.CpMode
	name    string
	kumactl *KumactlOptions
	cluster *UniversalCluster
	verbose bool
}

func NewUniversalControlPlane(t testing.TestingT, mode mode.CpMode, clusterName string, cluster *UniversalCluster, verbose bool) *UniversalControlPlane {
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

func (c *UniversalControlPlane) AddCluster(name, url, lbAddress string) error {
	return nil
}

func (c *UniversalControlPlane) GetKumaCPLogs() (string, error) {
	return "", nil
}

func (c *UniversalControlPlane) GetKDSServerAddress() string {
	return ""
}

func (c *UniversalControlPlane) GetGlobaStatusAPI() string {
	return ""
}
