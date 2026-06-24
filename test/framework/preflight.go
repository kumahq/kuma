package framework

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/onsi/ginkgo/v2"
)

// This file wires the external `kuma3-preflight` tool into the e2e run so each spec
// can be classified by the Kuma-3.0-deprecated features it exercises. It is inert
// unless both Kuma3PreflightBin (path to a built kuma3-preflight) and Kuma3PreflightDir
// (output directory) are configured (env KUMA3_PREFLIGHT_BIN / KUMA3_PREFLIGHT_DIR),
// so it adds nothing to a normal run.

// captureTimeoutBuffer is added to the tool's own --timeout to derive the hard context
// bound for a single capture: a wedged binary can never hang an AfterEach, yet the
// context never fires before the tool's own graceful stop.
const captureTimeoutBuffer = 15 * time.Second

var preflightSeq atomic.Int64

var preflightSlugRe = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// PreflightCaptureEnabled reports whether snapshot capture is configured. Callers
// guard on it before resolving the CP address so a normal run does no extra work.
func PreflightCaptureEnabled() bool {
	return Config != nil && Config.Kuma3PreflightBin != "" && Config.Kuma3PreflightDir != ""
}

// CapturePreflightCluster is the suite-facing entry point: it resolves the cluster's
// CP API address and captures a snapshot. It is a no-op unless PreflightCaptureEnabled(),
// and it never lets a capture problem fail a test — resolving the address is guarded
// (a K8s CP panics if its port-forward isn't up yet). For multizone, pass the global
// cluster: one audit of the global covers every zone (resources sync over KDS).
func CapturePreflightCluster(specName string, c Cluster) {
	if !PreflightCaptureEnabled() {
		return
	}
	addr := safeAPIServerAddress(c)
	if addr == "" {
		ginkgo.GinkgoWriter.Printf("preflight: no CP API address for %q, skipping\n", specName)
		return
	}
	CapturePreflight(specName, addr)
}

// safeAPIServerAddress returns the cluster's CP API address. It recovers from a panic
// (a K8s CP whose port-forward is not established panics in GetAPIServerAddress) and
// returns "" so capture stays best-effort and never fails a test.
func safeAPIServerAddress(c Cluster) string {
	defer func() {
		// Capture stays best-effort, but log the panic so a missing snapshot is
		// diagnosable instead of vanishing silently.
		if r := recover(); r != nil {
			ginkgo.GinkgoWriter.Printf("preflight: recovered while resolving CP API address: %v\n", r)
		}
	}()
	if c == nil {
		return ""
	}
	return c.GetKuma().GetAPIServerAddress()
}

// maxSlugLen bounds the human-readable slug embedded in a snapshot filename. Uniqueness
// comes from the p%02d-%05d- counter prefix, not the slug, so this only affects
// readability.
const maxSlugLen = 100

func preflightSlug(s string) string {
	s = strings.ToLower(strings.Trim(preflightSlugRe.ReplaceAllString(s, "-"), "-"))
	if s == "" {
		return "spec"
	}
	if len(s) > maxSlugLen {
		// FullText() prefixes every spec with its parent Describe/Context, so sibling
		// specs share a long common head and differ only in the tail (e.g. the
		// "-http"/"-tcp" variant). Keep both ends so the slug stays a useful identifier
		// instead of collapsing siblings to the same head-only prefix.
		head := maxSlugLen/2 - 1
		tail := maxSlugLen - head - 1
		s = strings.Trim(s[:head], "-") + "-" + strings.Trim(s[len(s)-tail:], "-")
	}
	return s
}

// CapturePreflight runs kuma3-preflight against the control plane at addr and writes a
// JSON snapshot. The filename carries the Ginkgo parallel-process index and a
// per-process sequence number, so concurrent processes sharing one output directory
// never collide (the classifier attributes by resource/mesh name, not file order, so
// ordering is not relied upon). It is a no-op unless PreflightCaptureEnabled(). A
// capture problem is logged but never fails the test: exit 1 (blockers found) and exit
// 3 (inconclusive) are expected outcomes and the report file is still written.
func CapturePreflight(specName, addr string) {
	if !PreflightCaptureEnabled() {
		return
	}
	bin := Config.Kuma3PreflightBin
	dir := Config.Kuma3PreflightDir
	toolTimeout := Config.Kuma3PreflightTimeout
	if err := os.MkdirAll(dir, 0o755); err != nil {
		ginkgo.GinkgoWriter.Printf("preflight: mkdir %s: %v\n", dir, err)
		return
	}
	seq := preflightSeq.Add(1)
	name := fmt.Sprintf("p%02d-%05d-%s.json", ginkgo.GinkgoParallelProcess(), seq, preflightSlug(specName))
	out := filepath.Join(dir, name)

	ctx, cancel := context.WithTimeout(context.Background(), toolTimeout+captureTimeoutBuffer)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "--address", addr, "--format", "json", "--output", out, "--timeout", toolTimeout.String())
	combined, err := cmd.CombinedOutput()
	if err == nil {
		return
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		switch ee.ExitCode() {
		case 1, 3: // blockers / inconclusive: expected, snapshot still written
			return
		}
	}
	// GinkgoWriter is only flushed on spec failure and capture never fails the spec, so
	// also emit a GitHub Actions ::warning:: to stdout: that keeps a broken nightly's
	// reason one click away instead of buried in passing-spec logs.
	ginkgo.GinkgoWriter.Printf("preflight: capture failed for %q: %v\n%s\n", specName, err, combined)
	fmt.Printf("::warning::preflight capture failed for %q: %v\n", specName, err)
}
