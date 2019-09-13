package envoy

import (
	"context"
	"io"
	"os/exec"

	"github.com/gogo/protobuf/proto"
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

type BootstrapConfigFactoryFunc func(cfg kuma_dp.Config) (proto.Message, error)

type Opts struct {
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

func (e *Envoy) Run(stop <-chan struct{}) error {
	bootstrapConfig, err := e.opts.Generator(e.opts.Config)
	if err != nil {
		return errors.Wrapf(err, "failed to generate Envoy bootstrap config")
	}
	configFile, err := newConfigFile(e.opts.Config.DataplaneRuntime, bootstrapConfig)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	args := []string{
		"-c", configFile,
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
	command := exec.CommandContext(ctx, e.opts.Config.DataplaneRuntime.BinaryPath, args...)
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
