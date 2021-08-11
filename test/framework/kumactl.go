package framework

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"
)

type KumactlOptions struct {
	t          testing.TestingT
	CPName     string
	Kumactl    string
	ConfigPath string
	Verbose    bool
	Env        map[string]string
}

func NewKumactlOptions(t testing.TestingT, cpname string, verbose bool) (*KumactlOptions, error) {
	kumactl := GetKumactlBin()

	_, err := os.Stat(kumactl)
	if kumactl == "" || os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "unable to find kumactl, please supply a valid KUMACTLBIN environment variable")
	}

	configPath := os.ExpandEnv(fmt.Sprintf(defaultKumactlConfig, cpname))

	return &KumactlOptions{
		t:          t,
		CPName:     cpname,
		Kumactl:    kumactl,
		ConfigPath: configPath,
		Verbose:    verbose,
		Env:        map[string]string{},
	}, nil
}

func (o *KumactlOptions) RunKumactl(args ...string) error {
	out, err := o.RunKumactlAndGetOutput(args...)
	if err != nil {
		return errors.Wrapf(err, out)
	}
	return nil
}

func (o *KumactlOptions) RunKumactlAndGetOutput(args ...string) (string, error) {
	return o.RunKumactlAndGetOutputV(o.Verbose, args...)
}

func (o *KumactlOptions) RunKumactlAndGetOutputV(verbose bool, args ...string) (string, error) {
	cmdArgs := []string{}
	if o.ConfigPath != "" {
		cmdArgs = append(cmdArgs, "--config-file", o.ConfigPath)
	}

	cmdArgs = append(cmdArgs, args...)
	command := shell.Command{
		Command: o.Kumactl,
		Args:    cmdArgs,
		Env:     o.Env,
	}

	if !verbose {
		command.Logger = logger.Discard
	}

	return shell.RunCommandAndGetStdOutE(o.t, command)
}

func (o *KumactlOptions) KumactlDelete(kumatype, name, mesh string) error {
	return o.RunKumactl("delete", kumatype, name, "--mesh", mesh)
}

func (o *KumactlOptions) KumactlList(kumatype, mesh string) ([]string, error) {
	out, err := o.RunKumactlAndGetOutput("get", kumatype, "--mesh", mesh, "-o", "json")
	if err != nil {
		return nil, err
	}

	type item struct {
		Name string `json:"name"`
	}
	type resourceList struct {
		Items []item `json:"items"`
	}

	list := &resourceList{}
	if err := json.Unmarshal([]byte(out), list); err != nil {
		return nil, err
	}

	var items []string
	for _, item := range list.Items {
		items = append(items, item.Name)
	}
	return items, nil
}

func (o *KumactlOptions) KumactlApply(configPath string) error {
	return o.RunKumactl("apply", "-f", configPath)
}

func (o *KumactlOptions) KumactlApplyFromString(configData string) error {
	tmpfile, err := storeConfigToTempFile(o.t.Name(), configData)
	if err != nil {
		return err
	}

	defer os.Remove(tmpfile)

	return o.KumactlApply(tmpfile)
}

func storeConfigToTempFile(name string, configData string) (string, error) {
	escapedTestName := url.PathEscape(name)

	tmpfile, err := ioutil.TempFile("", escapedTestName)
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	_, err = tmpfile.WriteString(configData)

	return tmpfile.Name(), err
}

func (o *KumactlOptions) KumactlInstallCP(mode string, args ...string) (string, error) {
	cmd := []string{
		"install", "control-plane",
	}

	cmd = append(cmd, "--mode", mode)
	switch mode {
	case core.Zone:
		cmd = append(cmd, "--zone", o.CPName)
		fallthrough
	case core.Global:
		if !UseLoadBalancer() {
			cmd = append(cmd, "--use-node-port")
		}
	}

	cmd = append(cmd, args...)

	return o.RunKumactlAndGetOutputV(
		false, // silence the log output of Install
		cmd...)
}

func (o *KumactlOptions) KumactlInstallDNS(args ...string) (string, error) {
	args = append([]string{"install", "dns"}, args...)

	return o.RunKumactlAndGetOutputV(
		false, // silence the log output of Install
		args...)
}

func (o *KumactlOptions) KumactlInstallMetrics() (string, error) {
	return o.RunKumactlAndGetOutput("install", "metrics")
}

func (o *KumactlOptions) KumactlInstallTracing() (string, error) {
	return o.RunKumactlAndGetOutput("install", "tracing")
}

func (o *KumactlOptions) KumactlConfigControlPlanesAdd(name, address string) error {
	_, err := retry.DoWithRetryE(o.t, "kumactl config control-planes add", DefaultRetries, DefaultTimeout,
		func() (string, error) {
			err := o.RunKumactl(
				"config", "control-planes", "add",
				"--overwrite",
				"--name", name,
				"--address", address)

			if err != nil {
				return "Unable to register Kuma CP. Try again.", err
			}

			return "", nil
		})

	return err
}
