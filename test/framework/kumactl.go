package framework

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/shell"
)

// KumactlOptions represents common options necessary to specify for all Kumactl calls
type KumactlOptions struct {
	testing.T
	CPName      string
	Kumactl     string
	ContextName string
	ConfigPath  string
	Verbose     bool
}

// NewKumactlOptions will return a pointer to new instance of KumactlOptions with the configured options
func NewKumactlOptions(cpname string, contextName string, configPath string, verbose bool) (*KumactlOptions, error) {
	kumactl := os.Getenv(envKUMACTLBIN)
	_, err := os.Stat(kumactl)
	if kumactl == "" || os.IsNotExist(err) {
		return nil, fmt.Errorf("Unable to find kumactl, please supply valid KUMACTL environment variable.")
	}

	if configPath == "" {
		configPath = os.ExpandEnv(fmt.Sprintf(defaultKumactlConfig, cpname))
	}

	return &KumactlOptions{
		CPName:      cpname,
		Kumactl:     kumactl,
		ContextName: contextName,
		ConfigPath:  configPath,
		Verbose:     verbose,
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
	}

	if !verbose {
		command.Logger = logger.Discard
	}
	return shell.RunCommandAndGetOutputE(o, command)
}

func (o *KumactlOptions) KumactlDelete(kumatype string, name string) error {
	return o.RunKumactl("delete", kumatype, name)
}

func (o *KumactlOptions) KumactlApply(configPath string) error {
	return o.RunKumactl("apply", "-f", configPath)
}

func (o *KumactlOptions) KumactlApplyFromString(configData string) error {
	tmpfile, err := storeConfigToTempFile(o.Name(), configData)
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

func (o *KumactlOptions) KumactlInstallCP() (string, error) {
	return o.RunKumactlAndGetOutputV(
		false, // silence the log output of Install
		"install", "control-plane",
		"--control-plane-image", kumaCPImage,
		"--dataplane-image", kumaDPImage,
		"--dataplane-init-image", kumaInitImage)
}

func (o *KumactlOptions) KumactlInstallMetrics() (string, error) {
	return o.RunKumactlAndGetOutput("install", "metrics")
}

func (o *KumactlOptions) KumactlInstallTracing() (string, error) {
	return o.RunKumactlAndGetOutput("install", "tracing")
}

func (o *KumactlOptions) KumactlConfigControlPlanesAdd(name, address string) error {
	return o.RunKumactl(
		"config", "control-planes", "add",
		"--name", name,
		"--address", address)
}
