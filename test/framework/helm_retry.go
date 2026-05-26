package framework

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
)

var transientHelmErrors = map[string]string{
	`(?i)\b(500|502|503|504)\b`:                "remote chart repository returned a transient server error",
	`(?i)\bEOF\b`:                              "remote chart download closed early",
	`(?i)connection reset by peer`:             "remote chart download connection reset",
	`(?i)connection refused`:                   "remote chart download connection refused",
	`(?i)i/o timeout`:                          "remote chart download timed out",
	`(?i)TLS handshake timeout`:                "remote chart download TLS handshake timed out",
	`(?i)Client\.Timeout exceeded`:             "remote chart download timed out",
	`(?i)no such host`:                         "remote chart repository DNS lookup failed",
	`(?i)temporary failure in name resolution`: "remote chart repository DNS lookup failed",
}

var helmChartPathRE = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func helmRunWithRetryE(t testing.TestingT, opts *helm.Options, args ...string) (string, error) {
	if len(args) == 0 {
		return "", retry.FatalError{Underlying: errors.New("helm command args are empty")}
	}

	command := args[0]
	commandArgs := args[1:]

	return retry.DoWithRetryableErrorsContextE(
		t,
		context.Background(),
		"helm "+strings.Join(args, " "),
		transientHelmErrors,
		5,
		10*time.Second,
		func() (string, error) {
			return helm.RunHelmCommandAndGetStdOutContextE(t, context.Background(), opts, command, commandArgs...)
		},
	)
}

func HelmUpgradeWithRetryE(t testing.TestingT, opts *helm.Options, chart, releaseName string) error {
	_, err := retry.DoWithRetryableErrorsContextE(
		t,
		context.Background(),
		"helm upgrade "+releaseName,
		transientHelmErrors,
		5,
		10*time.Second,
		func() (string, error) {
			return "", helm.UpgradeContextE(t, context.Background(), opts, chart, releaseName)
		},
	)
	return err
}

func HelmChartFromRepoE(t testing.TestingT, repoURL, chartName, version string) (string, error) {
	if repoURL == "" {
		return "", retry.FatalError{Underlying: errors.New("helm chart repo URL is empty")}
	}
	if chartName == "" {
		return "", retry.FatalError{Underlying: errors.New("helm chart name is empty")}
	}

	cacheRoot := os.Getenv("KUMA_E2E_HELM_CHART_CACHE")
	if cacheRoot == "" {
		cacheRoot = filepath.Join(os.TempDir(), "kuma-e2e-helm-chart-cache")
	}
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return "", err
	}

	versionKey := version
	if versionKey == "" {
		versionKey = "latest"
	}
	key := sha256.Sum256([]byte(repoURL + "\x00" + chartName + "\x00" + versionKey))
	cacheName := fmt.Sprintf(
		"%s-%s-%s.tgz",
		helmChartPathRE.ReplaceAllString(chartName, "-"),
		helmChartPathRE.ReplaceAllString(versionKey, "-"),
		hex.EncodeToString(key[:])[:12],
	)
	cachePath := filepath.Join(cacheRoot, cacheName)
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	tmpDir, err := os.MkdirTemp(cacheRoot, ".pull-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	args := []string{"pull", chartName, "--repo", repoURL, "--destination", tmpDir}
	if version != "" {
		args = append(args, "--version", version)
	}

	if _, err := helmRunWithRetryE(t, &helm.Options{}, args...); err != nil {
		return "", err
	}

	pulledCharts, err := filepath.Glob(filepath.Join(tmpDir, "*.tgz"))
	if err != nil {
		return "", err
	}
	if len(pulledCharts) != 1 {
		return "", errors.New("expected one pulled helm chart, got " + fmt.Sprint(len(pulledCharts)))
	}

	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if err := os.Rename(pulledCharts[0], cachePath); err != nil {
		if _, statErr := os.Stat(cachePath); statErr == nil {
			return cachePath, nil
		}
		return "", err
	}

	return cachePath, nil
}
