package builder

import (
	"bytes"
	"context"
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

// TODO(bartsmykla): Should we maybe log when verbose?
func findExecutable(nameParts ...string) executable {
	name := strings.Join(nameParts, "-")

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
	args []string
}

func newExecutable(name string, path string, args ...string) executable {
	return executable{
		name: name,
		path: path,
		args: args,
	}
}

func (e *executable) exec(ctx context.Context, args ...string) (*bytes.Buffer, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	// #nosec G204
	cmd := exec.CommandContext(ctx, e.path, append(e.args, args...)...)
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

// map[IPv6]string
var prefixes = map[bool]string{
	true:  "ip6tables",
	false: "iptables",
}

type executables struct {
	save                   executable
	restore                executable
	foundDockerOutputChain bool
}

func newExecutables(ipv6 bool, mode string) *executables {
	return &executables{
		save:    findExecutable(prefixes[ipv6], mode, "save"),
		restore: findExecutable(prefixes[ipv6], mode, "restore"),
	}
}

func (e *executables) verify(ctx context.Context, cfg config.Config) error {
	var missing []string

	if e.save.path == "" {
		missing = append(missing, e.save.name)
	}

	if e.restore.path == "" {
		missing = append(missing, e.restore.name)
	}

	if len(missing) > 0 {
		return errors.Errorf("couldn't find executables: [%s]", strings.Join(missing, ","))
	}

	// We always need to have access to the "nat" table
	if stdout, err := e.save.exec(ctx, "-t", "nat"); err != nil {
		return errors.Wrap(err, "couldn't verify if table: 'nat' is available")
	} else if cfg.ShouldRedirectDNS() || cfg.ShouldCaptureAllDNS() {
		e.foundDockerOutputChain = dockerOutputChainRegex.Match(stdout.Bytes())
	}

	if cfg.ShouldConntrackZoneSplit() {
		if _, err := e.save.exec(ctx, "-t", "raw"); err != nil {
			return errors.Wrap(err, "couldn't verify if table: 'raw' is available")
		}
	}

	return nil
}

func findIptables(ctx context.Context, cfg config.Config, ipv6 bool) (*executables, bool, error) {
	nft := newExecutables(ipv6, "nft")
	legacy := newExecutables(ipv6, "legacy")

	if err := nft.verify(ctx, cfg); err != nil {
		return legacy, true, legacy.verify(ctx, cfg)
	}

	// Found DOCKER_OUTPUT chain in iptables-nft
	if nft.foundDockerOutputChain {
		return nft, false, nil
	}

	if err := legacy.verify(ctx, cfg); err != nil {
		return nft, false, nil
	}

	// Found DOCKER_OUTPUT chain in iptables-legacy
	if legacy.foundDockerOutputChain {
		return legacy, true, nil
	}

	return nft, false, nil
}
