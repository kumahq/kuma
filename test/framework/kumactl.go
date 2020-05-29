package framework

import (
	"io/ioutil"
	"net/url"
	"os"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/shell"
	. "github.com/onsi/gomega"
)

func RunKumactl(t *TestFramework, options *KumactlOptions, args ...string) {
	err := RunKumactlE(t, options, args...)
	Expect(err).ToNot(HaveOccurred())
}

func RunKumactlE(t *TestFramework, options *KumactlOptions, args ...string) error {
	_, err := RunKumactlAndGetOutputE(t, options, args...)
	return err
}

func RunKumactlAndGetOutputE(t *TestFramework, options *KumactlOptions, args ...string) (string, error) {
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

func KumactlDelete(t *TestFramework, options *KumactlOptions, kumatype string, name string) {
	err := KumactlDeleteE(t, options, kumatype, name)
	Expect(err).ToNot(HaveOccurred())
}

func KumactlDeleteE(t *TestFramework, options *KumactlOptions, kumatype string, name string) error {
	return RunKumactlE(t, options, "delete", kumatype, name)
}

func KumactlApply(t *TestFramework, options *KumactlOptions, configPath string) {
	err := KumactlApplyE(t, options, configPath)
	Expect(err).ToNot(HaveOccurred())
}

func KumactlApplyE(t *TestFramework, options *KumactlOptions, configPath string) error {
	return RunKumactlE(t, options, "apply", "-f", configPath)
}

func KumactlApplyFromString(t *TestFramework, options *KumactlOptions, configData string) {
	err := KumactlApplyFromStringE(t, options, configData)
	Expect(err).ToNot(HaveOccurred())
}

func KumactlApplyFromStringE(t *TestFramework, options *KumactlOptions, configData string) error {
	tmpfile, err := StoreConfigToTempFileE(t, configData)
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile)
	return KumactlApplyE(t, options, tmpfile)
}

func StoreConfigToTempFile(t *TestFramework, configData string) string {
	out, err := StoreConfigToTempFileE(t, configData)
	Expect(err).ToNot(HaveOccurred())
	return out
}

func StoreConfigToTempFileE(t *TestFramework, configData string) (string, error) {
	escapedTestName := url.PathEscape(t.Name())
	tmpfile, err := ioutil.TempFile("", escapedTestName)
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	_, err = tmpfile.WriteString(configData)
	return tmpfile.Name(), err
}

func KumactlInstallCP(t *TestFramework, options *KumactlOptions) string {
	output, err := KumactlInstallCPE(t, options)
	Expect(err).ToNot(HaveOccurred())
	return output
}

func KumactlInstallCPE(t *TestFramework, options *KumactlOptions) (string, error) {
	return RunKumactlAndGetOutputE(t, options,
		"install", "control-plane",
		"--control-plane-image", kumaCPImage,
		"--dataplane-image", kumaDPImage,
		"--dataplane-init-image", kumaInitImage)
}

func KumactlInstallMetrics(t *TestFramework, options *KumactlOptions) string {
	output, err := KumactlInstallMetricsE(t, options)
	Expect(err).ToNot(HaveOccurred())
	return output
}

func KumactlInstallMetricsE(t *TestFramework, options *KumactlOptions) (string, error) {
	return RunKumactlAndGetOutputE(t, options, "install", "metrics")
}

func KumactlInstallTracing(t *TestFramework, options *KumactlOptions) string {
	output, err := KumactlInstallTracingE(t, options)
	Expect(err).ToNot(HaveOccurred())
	return output
}

func KumactlInstallTracingE(t *TestFramework, options *KumactlOptions) (string, error) {
	return RunKumactlAndGetOutputE(t, options, "install", "tracing")
}
