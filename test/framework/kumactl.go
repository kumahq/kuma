package framework

import (
	"fmt"
	"os"

	"github.com/gruntwork-io/terratest/modules/testing"

	"github.com/kumahq/kuma/test/framework/kumactl"
)

func NewKumactlOptionsE2E(t testing.TestingT, cpname string, verbose bool) *kumactl.KumactlOptions {
	configPath := os.ExpandEnv(fmt.Sprintf(defaultKumactlConfig, cpname))
	return kumactl.NewKumactlOptions(t, cpname, Config.KumactlBin, []string{}, configPath, verbose, DefaultRetries, DefaultTimeout)
}
