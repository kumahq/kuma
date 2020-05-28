package framework

import (
	"github.com/gruntwork-io/terratest/modules/logger"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/stretchr/testify/require"
)

func RunKumactl(t testing.TestingT, options *KumactlOptions, args ...string) {
	require.NoError(t, RunKumactlE(t, options, args...))
}

func RunKumactlE(t testing.TestingT, options *KumactlOptions, args ...string) error {
	_, err := RunKumactlAndGetOutputE(t, options, args...)
	return err
}

func RunKumactlAndGetOutputE(t testing.TestingT, options *KumactlOptions, args ...string) (string, error) {
	cmdArgs := []string{}
	if options.ConfigPath != "" {
		cmdArgs = append(cmdArgs, "--config-file", options.ConfigPath)
	}
	cmdArgs = append(cmdArgs, args...)
	command := shell.Command{
		Command: "kumactl",
		Args:    cmdArgs,
	}

	if options.Verbose == false {
		command.Logger = logger.Discard
	}
	return shell.RunCommandAndGetOutputE(t, command)
}

func KumactlDelete(t testing.TestingT, options *KumactlOptions, kumatype string, name string) {
	require.NoError(t, KumactlDeleteE(t, options, kumatype, name))
}

func KumactlDeleteE(t testing.TestingT, options *KumactlOptions, kumatype string, name string) error {
	return RunKumactlE(t, options, "delete", kumatype, name)
}

func KumactlApply(t testing.TestingT, options *KumactlOptions, configPath string) {
	require.NoError(t, KumactlApplyE(t, options, configPath))
}

func KumactlApplyE(t testing.TestingT, options *KumactlOptions, configPath string) error {
	return RunKumactlE(t, options, "apply", "-f", configPath)
}

func KumactlApplyFromString(t testing.TestingT, options *KumactlOptions, configData string) {
	require.NoError(t, KumactlApplyFromStringE(t, options, configData))
}

func KumactlApplyFromStringE(t testing.TestingT, options *KumactlOptions, configData string) error {
	tmpfile, err := StoreConfigToTempFileE(t, configData)
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile)
	return KumactlApplyE(t, options, tmpfile)
}

func StoreConfigToTempFile(t testing.TestingT, configData string) string {
	out, err := StoreConfigToTempFileE(t, configData)
	require.NoError(t, err)
	return out
}

func StoreConfigToTempFileE(t testing.TestingT, configData string) (string, error) {
	escapedTestName := url.PathEscape(t.Name())
	tmpfile, err := ioutil.TempFile("", escapedTestName)
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	_, err = tmpfile.WriteString(configData)
	return tmpfile.Name(), err
}

func KumactlInstallCP(t testing.TestingT, options *KumactlOptions) string {
	output, err := KumactlInstallCPE(t, options)
	require.NoError(t, err)
	return output
}

func KumactlInstallCPE(t testing.TestingT, options *KumactlOptions) (string, error) {
	return RunKumactlAndGetOutputE(t, options,
		"install", "control-plane",
		"--control-plane-image", "kuma/kuma-cp",
		"--dataplane-image", "kuma/kuma-dp",
		"--dataplane-init-image", "kuma/kuma-init")
}

func KumactlInstallMetrics(t testing.TestingT, options *KumactlOptions) string {
	output, err := KumactlInstallMetricsE(t, options)
	require.NoError(t, err)
	return output
}

func KumactlInstallMetricsE(t testing.TestingT, options *KumactlOptions) (string, error) {
	return RunKumactlAndGetOutputE(t, options, "install", "metrics")
}

func KumactlInstallTracing(t testing.TestingT, options *KumactlOptions) string {
	output, err := KumactlInstallTracingE(t, options)
	require.NoError(t, err)
	return output
}

func KumactlInstallTracingE(t testing.TestingT, options *KumactlOptions) (string, error) {
	return RunKumactlAndGetOutputE(t, options, "install", "tracing")
}
