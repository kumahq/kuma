package builder

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

var dockerOutputChainRegex = regexp.MustCompile(`(?m)^:DOCKER_OUTPUT`)

var fallbackPaths = []string{
	"/usr/sbin",
	"/sbin",
	"/usr/bin",
	"/bin",
}

func buildRestoreParameters(cfg config.Config, rulesFile *os.File, restoreLegacy bool) []string {
	return NewParameters().
		AppendIf(restoreLegacy, Wait(cfg.Wait), WaitInterval(cfg.WaitInterval)).
		Append(NoFlush()).
		Build(cfg.Verbose, rulesFile.Name())
}

func findExecutable(name string) executable {
	paths := append(
		[]string{name},
		fallbackPaths...,
	)

	for _, path := range paths {
		foundPath, err := exec.LookPath(path)
		if err == nil {
			return newExecutable(name, foundPath)
		}

		if errors.Is(err, exec.ErrDot) {
			if pwd, err := os.Getwd(); err == nil {
				return newExecutable(name, filepath.Join(pwd, foundPath))
			}

			return newExecutable(name, foundPath)
		}
	}

	return executable{name: name}
}

type executable struct {
	name string
	path string
}

func newExecutable(name string, path string) executable {
	return executable{
		name: name,
		path: path,
	}
}

func (e executable) exec(ctx context.Context, args ...string) (*bytes.Buffer, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	// #nosec G204
	cmd := exec.CommandContext(ctx, e.path, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return nil, errors.Wrap(err, stderr.String())
		}

		return nil, err
	}

	return &stdout, nil
}

type executables struct {
	iptables               executable
	save                   executable
	restore                executable
	fallback               *executables
	legacy                 bool
	foundDockerOutputChain bool
}

func newExecutables(ipv6 bool, mode string) *executables {
	prefix := iptables
	if ipv6 {
		prefix = ip6tables
	}

	iptables := fmt.Sprintf("%s-%s", prefix, mode)
	iptablesSave := fmt.Sprintf("%s-%s-%s", prefix, mode, "save")
	iptablesRestore := fmt.Sprintf("%s-%s-%s", prefix, mode, "restore")

	return &executables{
		iptables: findExecutable(iptables),
		save:     findExecutable(iptablesSave),
		restore:  findExecutable(iptablesRestore),
		legacy:   mode == "legacy",
	}
}

var necessaryMatchExtensions = []string{
	"owner",
	"tcp",
	"udp",
}

func (e *executables) verify(ctx context.Context, cfg config.Config) (*executables, error) {
	var missing []string

	if e.save.path == "" {
		missing = append(missing, e.save.name)
	}

	if e.restore.path == "" {
		missing = append(missing, e.restore.name)
	}

	if len(missing) > 0 {
		return nil, errors.Errorf("couldn't find executables: [%s]", strings.Join(missing, ","))
	}

	// We always need to have access to the "nat" table
	if stdout, err := e.save.exec(ctx, "-t", "nat"); err != nil {
		return nil, errors.Wrap(err, "couldn't verify if table: 'nat' is available")
	} else if cfg.ShouldRedirectDNS() || cfg.ShouldCaptureAllDNS() {
		e.foundDockerOutputChain = dockerOutputChainRegex.Match(stdout.Bytes())
	}

	// It seems in some cases (GKE with ContainerOS), even if "iptables-nft" is available
	// there are some kernel modules with iptables match extensions missing.
	for _, matchExtension := range necessaryMatchExtensions {
		if _, err := e.iptables.exec(ctx, "-m", matchExtension, "--help"); err != nil {
			return nil, errors.Wrapf(err, "verification if match: %q exist failed", matchExtension)
		}
	}

	if cfg.ShouldConntrackZoneSplit() {
		if _, err := e.save.exec(ctx, "-t", "raw"); err != nil {
			return nil, errors.Wrap(err, "couldn't verify if table: 'raw' is available")
		}
	}

	return e, nil
}

func (e *executables) withFallback(fallback *executables) *executables {
	if fallback != nil {
		e.fallback = fallback
	}

	return e
}

func detectIptablesExecutables(ctx context.Context, cfg config.Config, ipv6 bool) (*executables, error) {
	nft, nftVerifyErr := newExecutables(ipv6, "nft").verify(ctx, cfg)
	legacy, legacyVerifyErr := newExecutables(ipv6, "legacy").verify(ctx, cfg)

	if nftVerifyErr != nil && legacyVerifyErr != nil {
		return nil, fmt.Errorf("no valid iptables executable found: %s, %s", nftVerifyErr, legacyVerifyErr)
	}

	if nftVerifyErr != nil {
		return legacy, nil
	}

	// Found DOCKER_OUTPUT chain in iptables-nft
	if nft.foundDockerOutputChain {
		return nft.withFallback(legacy), nil
	}

	if legacyVerifyErr != nil {
		return nft, nil
	}

	// Found DOCKER_OUTPUT chain in iptables-legacy
	if legacy.foundDockerOutputChain {
		return legacy.withFallback(nft), nil
	}

	return nft.withFallback(legacy), nil
}
