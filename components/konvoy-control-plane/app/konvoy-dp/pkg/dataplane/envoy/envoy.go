package envoy

import (
	"context"
	"io"
	"os/exec"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	konvoy_dp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-dp"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
)

var (
	runLog = core.Log.WithName("konvoy-dataplane").WithName("run").WithName("envoy")
)

var (
	// overridable by unit tests
	newConfigFile = GenerateBootstrapFile
)

type BootstrapConfigFactoryFunc func(cfg konvoy_dp.Config) (proto.Message, error)

type Opts struct {
	Config    konvoy_dp.Config
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

	command := exec.CommandContext(ctx, e.opts.Config.DataplaneRuntime.BinaryPath, "-c", configFile)
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
