package framework

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
)

type KumactlOptions struct {
	t          testing.TestingT
	CPName     string
	Kumactl    string
	ConfigPath string
	Verbose    bool
	Env        map[string]string
}

func NewKumactlOptions(t testing.TestingT, cpname string, verbose bool) *KumactlOptions {
	configPath := os.ExpandEnv(fmt.Sprintf(defaultKumactlConfig, cpname))
	return &KumactlOptions{
		t:          t,
		CPName:     cpname,
		Kumactl:    Config.KumactlBin,
		ConfigPath: configPath,
		Verbose:    verbose,
		Env:        map[string]string{},
	}
}

func (k *KumactlOptions) RunKumactl(args ...string) error {
	out, err := k.RunKumactlAndGetOutput(args...)
	if err != nil {
		return errors.Wrapf(err, out)
	}
	return nil
}

func (k *KumactlOptions) RunKumactlAndGetOutput(args ...string) (string, error) {
	return k.RunKumactlAndGetOutputV(k.Verbose, args...)
}

func (k *KumactlOptions) RunKumactlAndGetOutputV(verbose bool, args ...string) (string, error) {
	cmdArgs := []string{}
	if k.ConfigPath != "" {
		cmdArgs = append(cmdArgs, "--config-file", k.ConfigPath)
	}

	cmdArgs = append(cmdArgs, args...)
	command := shell.Command{
		Command: k.Kumactl,
		Args:    cmdArgs,
		Env:     k.Env,
	}

	if !verbose {
		command.Logger = logger.Discard
	}

	return shell.RunCommandAndGetStdOutE(k.t, command)
}

func (k *KumactlOptions) KumactlDelete(kumatype, name, mesh string) error {
	return k.RunKumactl("delete", kumatype, name, "--mesh", mesh)
}

func (k *KumactlOptions) KumactlList(kumatype, mesh string) ([]string, error) {
	out, err := k.RunKumactlAndGetOutput("get", kumatype, "--mesh", mesh, "-o", "json")
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

func (k *KumactlOptions) KumactlApply(configPath string) error {
	return k.RunKumactl("apply", "-f", configPath)
}

func (k *KumactlOptions) KumactlApplyFromString(configData string) error {
	tmpfile, err := storeConfigToTempFile(k.t.Name(), configData)
	if err != nil {
		return err
	}

	defer os.Remove(tmpfile)

	return k.KumactlApply(tmpfile)
}

func storeConfigToTempFile(name string, configData string) (string, error) {
	escapedTestName := url.PathEscape(name)

	tmpfile, err := os.CreateTemp("", escapedTestName)
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	_, err = tmpfile.WriteString(configData)

	return tmpfile.Name(), err
}

func (k *KumactlOptions) KumactlInstallCP(mode string, args ...string) (string, error) {
	cmd := []string{
		"install", "control-plane",
	}

	cmd = append(cmd, "--mode", mode)
	switch mode {
	case core.Zone:
		cmd = append(cmd, "--zone", k.CPName)
		fallthrough
	case core.Global:
		if !Config.UseLoadBalancer {
			cmd = append(cmd, "--use-node-port")
		}
	}

	cmd = append(cmd, args...)

	return k.RunKumactlAndGetOutputV(
		false, // silence the log output of Install
		cmd...)
}

func (k *KumactlOptions) KumactlInstallObservability(namespace string, components []string) (string, error) {
	args := []string{"install", "observability", "--namespace", namespace}
	if len(components) != 0 {
		args = append(args, "--components", strings.Join(components, ","))
	}
	return k.RunKumactlAndGetOutput(args...)
}

func (k *KumactlOptions) KumactlConfigControlPlanesAdd(name, address, token string) error {
	_, err := retry.DoWithRetryE(k.t, "kumactl config control-planes add", DefaultRetries, DefaultTimeout,
		func() (string, error) {
			args := []string{
				"config", "control-planes", "add",
				"--overwrite",
				"--name", name,
				"--address", address,
			}
			if token != "" {
				args = append(args,
					"--auth-type", "tokens",
					"--auth-conf", "token="+token,
				)
			}
			err := k.RunKumactl(args...)

			if err != nil {
				return "Unable to register Kuma CP. Try again.", err
			}

			return "", nil
		})

	return err
}

// KumactlUpdateObject fetches an object and updates it after the update function is applied to it.
func (k *KumactlOptions) KumactlUpdateObject(
	typeName string,
	objectName string,
	update func(core_model.Resource) core_model.Resource,
) error {
	out, err := k.RunKumactlAndGetOutput("get", typeName, objectName, "-o", "yaml")
	if err != nil {
		return errors.Wrapf(err, "failed to get %q object %q", typeName, objectName)
	}

	resource, err := rest.YAML.UnmarshalCore([]byte(out))
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal %q object %q: %q", typeName, objectName, out)
	}

	updated := rest.From.Resource(update(resource))
	jsonRes, err := json.Marshal(updated)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal JSON for %q object %q", typeName, objectName)
	}

	return k.KumactlApplyFromString(string(jsonRes))
}
