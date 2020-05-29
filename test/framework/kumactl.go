package framework

import (
	"io/ioutil"
	"net/url"
	"os"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/shell"
)

func RunKumactl(t *TestFramework, options *KumactlOptions, args ...string) error {
	_, err := RunKumactlAndGetOutput(t, options, args...)
	return err
}

func RunKumactlAndGetOutput(t *TestFramework, options *KumactlOptions, args ...string) (string, error) {
	cmdArgs := []string{}
	if options.ConfigPath != "" {
		cmdArgs = append(cmdArgs, "--config-file", options.ConfigPath)
	}
	cmdArgs = append(cmdArgs, args...)
	command := shell.Command{
		Command: t.kumactl,
		Args:    cmdArgs,
	}

	if !options.Verbose {
		command.Logger = logger.Discard
	}
	return shell.RunCommandAndGetOutputE(t, command)
}

func KumactlDelete(t *TestFramework, options *KumactlOptions, kumatype string, name string) error {
	return RunKumactl(t, options, "delete", kumatype, name)
}

func KumactlApply(t *TestFramework, options *KumactlOptions, configPath string) error {
	return RunKumactl(t, options, "apply", "-f", configPath)
}

func KumactlApplyFromString(t *TestFramework, options *KumactlOptions, configData string) error {
	tmpfile, err := storeConfigToTempFile(t, configData)
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile)
	return KumactlApply(t, options, tmpfile)
}

func storeConfigToTempFile(t *TestFramework, configData string) (string, error) {
	escapedTestName := url.PathEscape(t.Name())
	tmpfile, err := ioutil.TempFile("", escapedTestName)
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	_, err = tmpfile.WriteString(configData)
	return tmpfile.Name(), err
}

func KumactlInstallCP(t *TestFramework, options *KumactlOptions) (string, error) {
	return RunKumactlAndGetOutput(t, options,
		"install", "control-plane",
		"--control-plane-image", kumaCPImage,
		"--dataplane-image", kumaDPImage,
		"--dataplane-init-image", kumaInitImage)
}

func KumactlInstallMetrics(t *TestFramework, options *KumactlOptions) (string, error) {
	return RunKumactlAndGetOutput(t, options, "install", "metrics")
}

func KumactlInstallTracing(t *TestFramework, options *KumactlOptions) (string, error) {
	return RunKumactlAndGetOutput(t, options, "install", "tracing")
}
