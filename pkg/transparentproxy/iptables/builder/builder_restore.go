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
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

var (
	dockerOutputChainRegex   = regexp.MustCompile(`(?m)^:DOCKER_OUTPUT`)
	fallbackPaths            = []string{"/usr/sbin", "/sbin", "/usr/bin", "/bin"}
	necessaryMatchExtensions = []string{"owner", "tcp", "udp"}
)

func buildRestoreParameters(cfg config.InitializedConfig, rulesFile *os.File, restoreLegacy bool) []string {
	return NewParameters().
		AppendIf(restoreLegacy, Wait(cfg.Wait), WaitInterval(cfg.WaitInterval)).
		Append(NoFlush()).
		Build(cfg.Verbose, rulesFile.Name())
}

func findExecutable(name string) Executable {
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

	return Executable{Name: name}
}

type Executable struct {
	Name string
	Path string
}

func newExecutable(name string, path string) Executable {
	return Executable{
		Name: name,
		Path: path,
	}
}

func (e Executable) exec(ctx context.Context, args ...string) (*bytes.Buffer, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	// #nosec G204
	cmd := exec.CommandContext(ctx, e.Path, args...)
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

type Executables struct {
	Iptables               Executable
	Save                   Executable
	Restore                Executable
	fallback               *Executables
	mode                   string
	foundDockerOutputChain bool
}

func newExecutables(ipv6 bool, mode string) *Executables {
	prefix := consts.Iptables
	if ipv6 {
		prefix = consts.Ip6tables
	}

	iptables := fmt.Sprintf("%s-%s", prefix, mode)
	iptablesSave := fmt.Sprintf("%s-%s-%s", prefix, mode, "save")
	iptablesRestore := fmt.Sprintf("%s-%s-%s", prefix, mode, "restore")

	return &Executables{
		Iptables: findExecutable(iptables),
		Save:     findExecutable(iptablesSave),
		Restore:  findExecutable(iptablesRestore),
		mode:     mode,
	}
}

func (e *Executables) legacy() bool {
	return e.mode == "legacy"
}

func (e *Executables) verify(ctx context.Context, cfg config.InitializedConfig) (*Executables, error) {
	var missing []string

	if e.Save.Path == "" {
		missing = append(missing, e.Save.Name)
	}

	if e.Restore.Path == "" {
		missing = append(missing, e.Restore.Name)
	}

	if len(missing) > 0 {
		return nil, errors.Errorf("couldn't find %s executables: [%s]", e.mode, strings.Join(missing, ", "))
	}

	// We always need to have access to the "nat" table
	if stdout, err := e.Save.exec(ctx, "-t", "nat"); err != nil {
		return nil, errors.Wrap(err, "couldn't verify if table: 'nat' is available")
	} else if cfg.ShouldRedirectDNS() || cfg.ShouldCaptureAllDNS() {
		e.foundDockerOutputChain = dockerOutputChainRegex.Match(stdout.Bytes())
	}

	// It seems in some cases (GKE with ContainerOS), even if "iptables-nft" is available
	// there are some kernel modules with iptables match extensions missing.
	for _, matchExtension := range necessaryMatchExtensions {
		if _, err := e.Iptables.exec(ctx, "-m", matchExtension, "--help"); err != nil {
			return nil, errors.Wrapf(err, "verification if match: %q exist failed", matchExtension)
		}
	}

	if cfg.ShouldConntrackZoneSplit(e.Iptables.Path) {
		if _, err := e.Save.exec(ctx, "-t", "raw"); err != nil {
			return nil, errors.Wrap(err, "couldn't verify if table: 'raw' is available")
		}
	}

	return e, nil
}

func (e *Executables) withFallback(fallback *Executables) *Executables {
	if fallback != nil {
		e.fallback = fallback
	}

	return e
}

func DetectIptablesExecutables(
	ctx context.Context,
	cfg config.InitializedConfig,
	ipv6 bool,
) (*Executables, error) {
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
