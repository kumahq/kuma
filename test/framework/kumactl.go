package framework

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/Kong/kuma/pkg/config/core"

	"github.com/pkg/errors"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/shell"
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
	kumactl := os.Getenv(envKUMACTLBIN)

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
	_, err := o.RunKumactlAndGetOutput(args...)
	return err
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

	return shell.RunCommandAndGetOutputE(o.t, command)
}

func (o *KumactlOptions) KumactlDelete(kumatype string, name string) error {
	return o.RunKumactl("delete", kumatype, name)
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

func (o *KumactlOptions) KumactlInstallCP(mode ...string) (string, error) {
	args := []string{
		"install", "control-plane",
		"--control-plane-image", kumaCPImage,
		"--dataplane-image", kumaDPImage,
		"--dataplane-init-image", kumaInitImage,
	}

	for _, m := range mode {
		args = append(args, "--mode", m)
		if m == core.Remote {
			args = append(args, "--cluster-name", o.CPName)
		}
	}

	return o.RunKumactlAndGetOutputV(
		false, // silence the log output of Install
		args...)
}

func (o *KumactlOptions) KumactlInstallDNS() (string, error) {
	args := []string{
		"install", "dns",
	}

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

func (o *KumactlOptions) KumactlInstallIngress() (string, error) {
	args := []string{
		"install", "ingress",
		"--image", kumaDPImage,
		"--use-node-port",
	}
	return o.RunKumactlAndGetOutput(args...)
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
				return "Unable to register Kuma CP. Try again.", fmt.Errorf("Unable to register Kuma CP. Try again.")
			}

			return "", nil
		})

	return err
}
