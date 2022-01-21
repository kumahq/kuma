package envoy

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	envoy_bootstrap_v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	"github.com/pkg/errors"

	command_utils "github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/command"
	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	pkg_log "github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/util/files"
)

var (
	runLog = core.Log.WithName("kuma-dp").WithName("run").WithName("envoy")
)

type BootstrapParams struct {
	Dataplane       *rest.Resource
	DNSPort         uint32
	EmptyDNSPort    uint32
	EnvoyVersion    EnvoyVersion
	DynamicMetadata map[string]string
}

type BootstrapConfigFactoryFunc func(url string, cfg kuma_dp.Config, params BootstrapParams) (*envoy_bootstrap_v3.Bootstrap, []byte, error)

type Opts struct {
	Config          kuma_dp.Config
	BootstrapConfig []byte
	Dataplane       *rest.Resource
	Stdout          io.Writer
	Stderr          io.Writer
	Quit            chan struct{}
	LogLevel        pkg_log.LogLevel
}

func New(opts Opts) (*Envoy, error) {
	if _, err := lookupEnvoyPath(opts.Config.DataplaneRuntime.BinaryPath); err != nil {
		runLog.Error(err, "could not find the envoy executable in your path")
		return nil, err
	}
	return &Envoy{opts: opts}, nil
}

var _ component.Component = &Envoy{}

type Envoy struct {
	opts Opts
}

type EnvoyVersion struct {
	Build   string
	Version string
}

func (e *Envoy) NeedLeaderElection() bool {
	return false
}

func lookupEnvoyPath(configuredPath string) (string, error) {
	return files.LookupBinaryPath(
		files.LookupInPath(configuredPath),
		files.LookupInCurrentDirectory("envoy"),
		files.LookupNextToCurrentExecutable("envoy"),
	)
}

func (e *Envoy) Start(stop <-chan struct{}) error {
	configFile, err := GenerateBootstrapFile(e.opts.Config.DataplaneRuntime, e.opts.BootstrapConfig)
	if err != nil {
		return err
	}
	runLog.Info("bootstrap configuration saved to a file", "file", configFile)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	binaryPathConfig := e.opts.Config.DataplaneRuntime.BinaryPath
	resolvedPath, err := lookupEnvoyPath(binaryPathConfig)
	if err != nil {
		return err
	}

	args := []string{
		"--config-path", configFile,
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
		"--log-level", e.opts.LogLevel.String(),
	}

	// If the concurrency is explicit, use that. On Linux, users
	// can also implicitly set concurrency using cpusets.
	if e.opts.Config.DataplaneRuntime.Concurrency > 0 {
		args = append(args,
			"--concurrency",
			strconv.FormatUint(uint64(e.opts.Config.DataplaneRuntime.Concurrency), 10),
		)
	} else if runtime.GOOS == "linux" {
		// The `--cpuset-threads` flag is still present on
		// non-Linux, but emits a warning that we might as well
		// avoid.
		args = append(args, "--cpuset-threads")
	}

	command := command_utils.BuildCommand(ctx, e.opts.Stdout, e.opts.Stderr, resolvedPath, args...)

	runLog.Info("starting Envoy", "path", resolvedPath, "arguments", args)
	if err := command.Start(); err != nil {
		runLog.Error(err, "envoy executable failed", "path", resolvedPath, "arguments", args)
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- command.Wait()
	}()

	select {
	case <-stop:
		runLog.Info("stopping Envoy")
		cancel()
		return nil
	case err := <-done:
		if err != nil {
			runLog.Error(err, "Envoy terminated with an error")
		} else {
			runLog.Info("Envoy terminated successfully")
		}
		if e.opts.Quit != nil {
			close(e.opts.Quit)
		}

		return err
	}
}

func GetEnvoyVersion(binaryPath string) (*EnvoyVersion, error) {
	resolvedPath, err := lookupEnvoyPath(binaryPath)
	if err != nil {
		return nil, err
	}
	arg := "--version"
	command := exec.Command(resolvedPath, arg)
	output, err := command.Output()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute %s with arguments %q", resolvedPath, arg)
	}
	build := strings.ReplaceAll(string(output), "\r\n", "\n")
	build = strings.Trim(build, "\n")
	build = regexp.MustCompile(`version:(.*)`).FindString(build)
	build = strings.Trim(build, "version:")
	build = strings.Trim(build, " ")

	parts := strings.Split(build, "/")
	if len(parts) != 5 { // revision/build_version_number/revision_status/build_type/ssl_version
		return nil, errors.Errorf("wrong Envoy build format: %s", build)
	}
	return &EnvoyVersion{
		Build:   build,
		Version: parts[1],
	}, nil
}
