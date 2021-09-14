package dnsserver

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"

	command_utils "github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/command"
	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core"
)

var (
	runLog = core.Log.WithName("kuma-dp").WithName("run").WithName("dns-server")
)

type DNSServer struct {
	opts *Opts
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

func lookupDNSServerPath(configuredPath string) (string, error) {
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
		selfPath + "/coredns",
		cwd + "/coredns",
	})
	if err != nil {
		return "", err
	}

	return path, nil
}

func New(opts *Opts) (*DNSServer, error) {
	if _, err := lookupDNSServerPath(opts.Config.DNS.CoreDNSBinaryPath); err != nil {
		runLog.Error(err, "could not find the DNS Server executable in your path")
		return nil, err
	}

	return &DNSServer{opts: opts}, nil
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

	binaryPathConfig := dnsConfig.CoreDNSBinaryPath
	resolvedPath, err := lookupDNSServerPath(binaryPathConfig)
	if err != nil {
		return err
	}

	args := []string{
		"-conf", configFile,
		"-quiet",
	}

	command := command_utils.BuildCommand(ctx, s.opts.Stdout, s.opts.Stderr, resolvedPath, args...)

	runLog.Info("starting DNS Server (coredns)", "args", args)

	if err := command.Start(); err != nil {
		runLog.Error(err, "the DNS Server executable was found at "+resolvedPath+" but an error occurred when executing it")
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
