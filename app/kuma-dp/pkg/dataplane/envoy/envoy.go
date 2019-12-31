package envoy

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/Kong/kuma/pkg/catalog"

	"path/filepath"

	"github.com/golang/protobuf/proto"

	"github.com/pkg/errors"

	kuma_dp "github.com/Kong/kuma/pkg/config/app/kuma-dp"
	"github.com/Kong/kuma/pkg/core"
)

var (
	runLog = core.Log.WithName("kuma-dp").WithName("run").WithName("envoy")
)

var (
	// overridable by unit tests
	newConfigFile = GenerateBootstrapFile
)

type BootstrapConfigFactoryFunc func(url string, cfg kuma_dp.Config) (proto.Message, error)

type Opts struct {
	Catalog   catalog.Catalog
	Config    kuma_dp.Config
	Generator BootstrapConfigFactoryFunc
	Stdout    io.Writer
	Stderr    io.Writer
}

func New(opts Opts) *Envoy {
	return &Envoy{opts: opts}
}

type Envoy struct {
	opts Opts
}

func getSelfPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Dir(ex), nil
}

func lookupBinaryPath(candidatePaths []string) (string, error) {
	for _, candidatePath := range candidatePaths {
		path, err := exec.LookPath(candidatePath)
		if err == nil {
			return path, nil
		}
	}

	return "", errors.Errorf("could not find binary in any of the following paths: %v", candidatePaths)
}

func lookupEnvoyPath(configuredPath string) (string, error) {
	selfPath, err := getSelfPath()
	if err != nil {
		return "", err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	path, err := lookupBinaryPath([]string{
		configuredPath,
		selfPath + "/envoy",
		cwd + "/envoy",
	})
	if err != nil {
		return "", err
	}

	return path, nil
}

func (e *Envoy) Run(stop <-chan struct{}) error {
	bootstrapConfig, err := e.opts.Generator(e.opts.Catalog.Apis.Bootstrap.Url, e.opts.Config)
	if err != nil {
		return errors.Wrapf(err, "failed to generate Envoy bootstrap config")
	}
	configFile, err := newConfigFile(e.opts.Config.DataplaneRuntime, bootstrapConfig)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	binaryPathConfig := e.opts.Config.DataplaneRuntime.BinaryPath
	resolvedPath, err := lookupEnvoyPath(binaryPathConfig)
	if err != nil {
		return err
	}

	args := []string{
		"-c", configFile,
		"--drain-time-s",
		fmt.Sprintf("%d", e.opts.Config.Dataplane.DrainTime/time.Second),
		// "hot restart" (enabled by default) requires each Envoy instance to have
		// `--base-id <uint32_t>` argument.
		// it is not possible to start multiple Envoy instances on the same Linux machine
		// without `--base-id <uint32_t>` set.
		// although we could come up with a solution how to generate `--base-id <uint32_t>`
		// automatically, it is not strictly necessary since we're not using "hot restart"
		// and we don't expect users to do "hot restart" manually.
		// so, let's turn it off to simplify getting started experience.
		"--disable-hot-restart",
	}
	command := exec.CommandContext(ctx, resolvedPath, args...)
	command.Stdout = e.opts.Stdout
	command.Stderr = e.opts.Stderr
	if err := command.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- command.Wait()
	}()

	select {
	case <-stop:
		cancel()
		return nil
	case err := <-done:
		if err != nil {
			runLog.Error(err, "Envoy terminated with an error")
		} else {
			runLog.Info("Envoy terminated successfully")
		}
		return err
	}
}
