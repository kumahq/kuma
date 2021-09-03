package envoy

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	command_utils "github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/command"
	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	pkg_log "github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
)

var (
	runLog = core.Log.WithName("kuma-dp").WithName("run").WithName("envoy")
)

type BootstrapParams struct {
	Dataplane        *rest.Resource
	BootstrapVersion types.BootstrapVersion
	DNSPort          uint32
	EmptyDNSPort     uint32
	EnvoyVersion     EnvoyVersion
	DynamicMetadata  map[string]string
}

type BootstrapConfigFactoryFunc func(url string, cfg kuma_dp.Config, params BootstrapParams) ([]byte, types.BootstrapVersion, error)

type Opts struct {
	Config          kuma_dp.Config
	Generator       BootstrapConfigFactoryFunc
	Dataplane       *rest.Resource
	DynamicMetadata map[string]string
	DNSPort         uint32
	EmptyDNSPort    uint32
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

func (e *Envoy) Start(stop <-chan struct{}) error {
	envoyVersion, err := e.version()
	if err != nil {
		return errors.Wrap(err, "failed to get Envoy version")
	}
	runLog.Info("fetched Envoy version", "version", envoyVersion)
	runLog.Info("generating bootstrap configuration")
	bootstrapConfig, version, err := e.opts.Generator(e.opts.Config.ControlPlane.URL, e.opts.Config, BootstrapParams{
		Dataplane:        e.opts.Dataplane,
		BootstrapVersion: types.BootstrapVersion(e.opts.Config.Dataplane.BootstrapVersion),
		DNSPort:          e.opts.DNSPort,
		EmptyDNSPort:     e.opts.EmptyDNSPort,
		EnvoyVersion:     *envoyVersion,
		DynamicMetadata:  e.opts.DynamicMetadata,
	})
	if err != nil {
		return errors.Errorf("Failed to generate Envoy bootstrap config. %v", err)
	}
	configFile, err := GenerateBootstrapFile(e.opts.Config.DataplaneRuntime, bootstrapConfig)
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
	if version != "" { // version is always send by Kuma CP, but we check empty for backwards compatibility reasons (new Kuma DP connects to old Kuma CP)
		args = append(args, "--bootstrap-version", string(version))
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

	runLog.Info("starting Envoy", "args", args)
	if err := command.Start(); err != nil {
		runLog.Error(err, "the envoy executable was found at "+resolvedPath+" but an error occurred when executing it")
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

func (e *Envoy) version() (*EnvoyVersion, error) {
	binaryPathConfig := e.opts.Config.DataplaneRuntime.BinaryPath
	resolvedPath, err := lookupEnvoyPath(binaryPathConfig)
	if err != nil {
		return nil, err
	}
	arg := "--version"
	command := exec.Command(resolvedPath, arg)
	output, err := command.Output()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("the envoy excutable was found at %s but an error occurred when executing it with arg %s", resolvedPath, arg))
	}
	build := strings.ReplaceAll(string(output), "\r\n", "\n")
	build = strings.Trim(build, "\n")
	build = regexp.MustCompile(`version:(.*)`).FindString(build)
	build = strings.Trim(build, "version:")
	build = strings.Trim(build, " ")
	version := regexp.MustCompile(`/([0-9.]+)/`).FindString(build)
	version = strings.Trim(version, "/")
	return &EnvoyVersion{
		Build:   build,
		Version: version,
	}, nil
}
