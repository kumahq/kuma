package dnsserver

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"regexp"
	"sync"
	"text/template"

	"github.com/pkg/errors"

	command_utils "github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/command"
	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/util/files"
)

var runLog = core.Log.WithName("kuma-dp").WithName("run").WithName("dns-server")

type DNSServer struct {
	opts *Opts
	path string

	wg sync.WaitGroup
}

var _ component.GracefulComponent = &DNSServer{}

type Opts struct {
	Config                   kuma_dp.Config
	Stdout                   io.Writer
	Stderr                   io.Writer
	OnFinish                 context.CancelFunc
	ProvidedCorefileTemplate []byte
}

func lookupDNSServerPath(configuredPath string) (string, error) {
	return files.LookupBinaryPath(
		files.LookupInPath(configuredPath),
		files.LookupInCurrentDirectory("coredns"),
		files.LookupNextToCurrentExecutable("coredns"),
	)
}

func New(opts *Opts) (*DNSServer, error) {
	dnsServerPath, err := lookupDNSServerPath(opts.Config.DNS.CoreDNSBinaryPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not find coreDNS executable")
	}
	if opts.OnFinish == nil {
		opts.OnFinish = func() {}
	}

	return &DNSServer{opts: opts, path: dnsServerPath}, nil
}

func (s *DNSServer) GetVersion() (string, error) {
	path := s.path
	command := exec.Command(path, "--version")
	output, err := command.Output()
	if err != nil {
		return "", errors.Wrapf(err, "failed to execute coreDNS at path %s", path)
	}

	match := regexp.MustCompile(`CoreDNS-(.*)`).FindSubmatch(output)
	if len(match) < 2 {
		return "", errors.Errorf("unexpected version output format: %s", output)
	}

	return string(match[1]), nil
}

func (s *DNSServer) NeedLeaderElection() bool {
	return false
}

func (s *DNSServer) Start(stop <-chan struct{}) error {
	s.wg.Add(1)
	// Component should only be considered done after CoreDNS exists.
	// Otherwise, we may not propagate SIGTERM on time.
	defer func() {
		s.wg.Done()
		s.opts.OnFinish()
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dnsConfig := s.opts.Config.DNS

	configFile, err := s.generateCoreFile(dnsConfig)
	if err != nil {
		return err
	}

	args := []string{
		"-conf", configFile,
		"-quiet",
	}

	command := command_utils.BuildCommand(ctx, s.opts.Stdout, s.opts.Stderr, s.path, args...)

	runLog.Info("starting DNS Server (coredns)", "args", args)

	if err := command.Start(); err != nil {
		runLog.Error(err, "the DNS Server executable was found at "+s.path+" but an error occurred when executing it")
		return err
	}

	go func() {
		<-stop
		runLog.Info("stopping DNS Server")
		cancel()
	}()
	err = command.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		runLog.Error(err, "DNS Server terminated with an error")
		return err
	}
	runLog.Info("DNS Server terminated successfully")
	return nil
}

func (s *DNSServer) generateCoreFile(dnsConfig kuma_dp.DNS) (string, error) {
	var tmpl *template.Template

	if dnsConfig.CoreDNSConfigTemplatePath != "" {
		t, err := template.ParseFiles(dnsConfig.CoreDNSConfigTemplatePath)
		if err != nil {
			return "", err
		}

		tmpl = t
	} else {
		var corefileTemplate []byte
		if len(s.opts.ProvidedCorefileTemplate) > 0 {
			corefileTemplate = s.opts.ProvidedCorefileTemplate
		} else {
			embedded, err := config.ReadFile("Corefile")
			if err != nil {
				return "", errors.Wrap(err, "couldn't open embedded Corefile")
			}
			corefileTemplate = embedded
		}

		t, err := template.New("Corefile").Parse(string(corefileTemplate))
		if err != nil {
			return "", err
		}

		tmpl = t
	}

	bs := bytes.NewBuffer([]byte{})

	if err := tmpl.Execute(bs, dnsConfig); err != nil {
		return "", err
	}

	configFile, err := WriteCorefile(dnsConfig, bs.Bytes())
	if err != nil {
		return "", err
	}
	runLog.Info("configuration saved to a file", "file", configFile)
	return configFile, nil
}

func (s *DNSServer) WaitForDone() {
	s.wg.Wait()
}
