package dnsserver

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"regexp"
	"text/template"

	"github.com/pkg/errors"

	command_utils "github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/command"
	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/util/files"
)

var (
	runLog = core.Log.WithName("kuma-dp").WithName("run").WithName("dns-server")
)

type DNSServer struct {
	opts *Opts
	path string
}

type Opts struct {
	Config kuma_dp.Config
	Stdout io.Writer
	Stderr io.Writer
	Quit   chan struct{}
}

// DefaultCoreFileTemplate defines the template to use to configure coreDNS to use the envoy dns filter.
const DefaultCoreFileTemplate = `.:{{ .CoreDNSPort }} {
    forward . 127.0.0.1:{{ .EnvoyDNSPort }}
    # We want all requests to be sent to the Envoy DNS Filter, unsuccessful responses should be forwarded to the original DNS server.
    # For example: requests other than A, AAAA and SRV will return NOTIMP when hitting the envoy filter and should be sent to the original DNS server.
    # Codes from: https://github.com/miekg/dns/blob/master/msg.go#L138
    alternate NOTIMP,FORMERR,NXDOMAIN,SERVFAIL,REFUSED . /etc/resolv.conf
    prometheus localhost:{{ .PrometheusPort }}
    errors
}

.:{{ .CoreDNSEmptyPort }} {
    template ANY ANY . {
      rcode NXDOMAIN
    }
}`

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
		runLog.Error(err, "could not find the DNS Server executable in your path")
		return nil, err
	}

	return &DNSServer{opts: opts, path: dnsServerPath}, nil
}

func (s *DNSServer) GetVersion() (string, error) {
	command := exec.Command(s.path, "--version")
	output, err := command.Output()
	if err != nil {
		return "", err
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dnsConfig := s.opts.Config.DNS

	var tmpl *template.Template

	if dnsConfig.CoreDNSConfigTemplatePath != "" {
		t, err := template.ParseFiles(dnsConfig.CoreDNSConfigTemplatePath)
		if err != nil {
			return err
		}

		tmpl = t
	} else {
		t, err := template.New("Corefile").Parse(DefaultCoreFileTemplate)
		if err != nil {
			return err
		}

		tmpl = t
	}

	bs := bytes.NewBuffer([]byte{})

	if err := tmpl.Execute(bs, dnsConfig); err != nil {
		return err
	}

	configFile, err := GenerateConfigFile(dnsConfig, bs.Bytes())
	if err != nil {
		return err
	}
	runLog.Info("configuration saved to a file", "file", configFile)

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

	done := make(chan error, 1)

	go func() {
		done <- command.Wait()
	}()

	select {
	case <-stop:
		runLog.Info("stopping DNS Server")
		cancel()
		return nil
	case err := <-done:
		if err != nil {
			runLog.Error(err, "DNS Server terminated with an error")
		} else {
			runLog.Info("DNS Server terminated successfully")
		}

		if s.opts.Quit != nil {
			close(s.opts.Quit)
		}

		return err
	}
}
