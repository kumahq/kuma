package framework

import (
	"strconv"

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

func (c *UniversalControlPlane) AddCluster(name, lbAddress, kdsAddress, ingressAddress string) error {
	cat := NewSshApp(false, c.cluster.apps[AppModeCP].ports["22"], []string{}, []string{
		"cat", confPath,
	})
	err := cat.Run()
	if err != nil {
		return err
	}

	//time.Sleep(time.Second)

	resultYAML, err := addGlobal(cat.Out(), lbAddress, kdsAddress, ingressAddress)
	if err != nil {
		return err
	}

	err = NewSshApp(false, c.cluster.apps[AppModeCP].ports["22"], []string{}, []string{
		"echo", "\"" + resultYAML + "\"", ">", confPath,
	}).Run()

	return err
}

func (c *UniversalControlPlane) GetKumaCPLogs() (string, error) {
	panic("not implemented")
}

func (c *UniversalControlPlane) GetKDSServerAddress() string {
	return "grpcs://" + c.cluster.apps[AppModeCP].ip + ":5685"
}

func (c *UniversalControlPlane) GetIngressAddress() string {
	return c.cluster.apps[AppModeCP].ip + ":" + strconv.FormatUint(uint64(kdsPort), 10)
}

func (c *UniversalControlPlane) GetGlobaStatusAPI() string {
	panic("not implemented")
}
