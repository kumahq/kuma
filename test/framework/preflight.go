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
// unless both KUMA3_PREFLIGHT_BIN (path to a built kuma3-preflight) and
// KUMA3_PREFLIGHT_DIR (output directory) are set, so it adds nothing to a normal run.

// captureTimeout hard-bounds a single capture so a wedged binary can never hang an
// AfterEach. It is larger than the tool's own --timeout below, which is the normal stop.
const captureTimeout = 60 * time.Second

var preflightSeq atomic.Int64

var preflightSlugRe = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// PreflightCaptureEnabled reports whether snapshot capture is configured. Callers
// guard on it before resolving the CP address so a normal run does no extra work.
func PreflightCaptureEnabled() bool {
	return os.Getenv("KUMA3_PREFLIGHT_BIN") != "" && os.Getenv("KUMA3_PREFLIGHT_DIR") != ""
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

func preflightSlug(s string) string {
	s = strings.ToLower(strings.Trim(preflightSlugRe.ReplaceAllString(s, "-"), "-"))
	if len(s) > 80 {
		s = s[:80]
	}
	if s == "" {
		s = "spec"
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
	bin := os.Getenv("KUMA3_PREFLIGHT_BIN")
	dir := os.Getenv("KUMA3_PREFLIGHT_DIR")
	if bin == "" || dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		ginkgo.GinkgoWriter.Printf("preflight: mkdir %s: %v\n", dir, err)
		return
	}
	seq := preflightSeq.Add(1)
	name := fmt.Sprintf("p%02d-%05d-%s.json", ginkgo.GinkgoParallelProcess(), seq, preflightSlug(specName))
	out := filepath.Join(dir, name)

	ctx, cancel := context.WithTimeout(context.Background(), captureTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "--address", addr, "--format", "json", "--output", out, "--timeout", "45s")
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
	ginkgo.GinkgoWriter.Printf("preflight: capture failed for %q: %v\n%s\n", specName, err, combined)
}
