package framework

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
)

const (
	externalServiceLowPort = 10204
)

func NewExernalService(t testing.TestingT, clusterName string, mode AppMode, args []string) (*UniversalApp, error) {
	app := &UniversalApp{
		t:            t,
		ports:        map[string]string{},
		lastUsedPort: externalServiceLowPort,
		verbose:      false,
	}

	app.allocatePublicPortsFor("22", "80")

	opts := defaultDockerOptions
	opts.OtherOptions = append(opts.OtherOptions, "--name", clusterName+"_"+string(mode))
	opts.OtherOptions = append(opts.OtherOptions, "--network", "kind")
	opts.OtherOptions = append(opts.OtherOptions, app.publishPortsForDocker()...)

	container, err := docker.RunAndGetIDE(t, kumaUniversalImage, &opts)
	if err != nil {
		return nil, err
	}

	app.container = container

	retry.DoWithRetry(app.t, "get IP "+app.container, DefaultRetries, DefaultTimeout,
		func() (string, error) {
			app.ip, err = app.getIP()
			if err != nil {
				return "Unable to get Container IP", err
			}
			return "Success", nil
		})

	fmt.Printf("External Service IP %s\n", app.ip)

	app.CreateMainApp([]string{}, args)

	return app, nil
}
