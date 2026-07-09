package utils_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/test/framework/utils"
)

const (
	hashA = "a1b2c3d4e5f60718"
	hashB = "b1b2c3d4e5f60718"
	hashC = "c1b2c3d4e5f60718"
)

// changedLine renders a "config has changed" log line the way the console
// encoder does: extra fields as trailing compact JSON.
func changedLine(proxyName string, hashes ...string) string {
	quoted := make([]string, len(hashes))
	for i, h := range hashes {
		quoted[i] = fmt.Sprintf("%q", h)
	}
	return fmt.Sprintf(
		`2026-07-08T00:00:00Z	INFO	xds.reconcile	config has changed	{"proxyName": %q, "mesh": "default", "versions": [%s]}`,
		proxyName, strings.Join(quoted, ", "),
	)
}

// changedLineLogfmt renders a "config has changed" log line using a
// logfmt-style proxyName=value field, as a fallback encoding.
func changedLineLogfmt(proxyName string, hashes ...string) string {
	quoted := make([]string, len(hashes))
	for i, h := range hashes {
		quoted[i] = fmt.Sprintf("%q", h)
	}
	return fmt.Sprintf(
		`config has changed proxyName=%s mesh=default "versions": [%s]`,
		proxyName, strings.Join(quoted, ", "),
	)
}

// changedLineForDemoClient renders a console-encoded "config has changed" line
// for a demo-client proxy in a specific mesh, so tests can exercise proxies
// that share a name across meshes.
func changedLineForDemoClient(mesh string, hashes ...string) string {
	quoted := make([]string, len(hashes))
	for i, h := range hashes {
		quoted[i] = fmt.Sprintf("%q", h)
	}
	return fmt.Sprintf(
		`2026-07-08T00:00:00Z	INFO	xds.reconcile	config has changed	{"proxyName": "demo-client", "mesh": %q, "versions": [%s]}`,
		mesh, strings.Join(quoted, ", "),
	)
}

func joinLines(lines ...string) string {
	return strings.Join(lines, "\n")
}

var _ = Describe("DetectXdsChurn", func() {
	It("flags a proxy whose config regenerates the same hash 3+ times", func() {
		logs := joinLines(
			changedLine("backend", hashA),
			changedLine("backend", hashA),
			changedLine("backend", hashA),
		)

		Expect(utils.DetectXdsChurn(logs)).To(ConsistOf(
			"proxy backend in mesh default regenerated identical config 3 times (hash a1b2c3d4e5f60718) — non-deterministic xDS",
		))
	})

	It("does not flag a monotonic config change or a single revert (A -> B -> A)", func() {
		logs := joinLines(
			changedLine("backend", hashA),
			changedLine("backend", hashB),
			changedLine("backend", hashA),
		)

		Expect(utils.DetectXdsChurn(logs)).To(BeEmpty())
	})

	It("does not flag non-consecutive reappearances of the same hash", func() {
		logs := joinLines(
			changedLine("backend", hashA),
			changedLine("backend", hashA),
			changedLine("backend", hashB),
			changedLine("backend", hashA),
			changedLine("backend", hashA),
		)

		Expect(utils.DetectXdsChurn(logs)).To(BeEmpty())
	})

	It("only flags the proxy that is actually churning", func() {
		logs := joinLines(
			changedLine("backend", hashA),
			changedLine("backend", hashA),
			changedLine("backend", hashA),
			changedLine("frontend", hashA),
			changedLine("frontend", hashB),
		)

		Expect(utils.DetectXdsChurn(logs)).To(ConsistOf(
			"proxy backend in mesh default regenerated identical config 3 times (hash a1b2c3d4e5f60718) — non-deterministic xDS",
		))
	})

	It("parses the proxyName=value fallback format", func() {
		logs := joinLines(
			changedLineLogfmt("backend", hashA),
			changedLineLogfmt("backend", hashA),
			changedLineLogfmt("backend", hashA),
		)

		Expect(utils.DetectXdsChurn(logs)).To(ConsistOf(
			"proxy backend in mesh default regenerated identical config 3 times (hash a1b2c3d4e5f60718) — non-deterministic xDS",
		))
	})

	It("ignores 16-hex tokens that appear outside the versions array", func() {
		line := `2026-07-08T00:00:00Z	INFO	xds.reconcile	config has changed	{"proxyName": "backend", "requestId": "deadbeefdeadbeef", "versions": []}`
		logs := joinLines(line, line, line)

		Expect(utils.DetectXdsChurn(logs)).To(BeEmpty())
	})

	It("counts multiple hashes in one versions array independently", func() {
		logs := joinLines(
			changedLine("backend", hashA, hashB),
			changedLine("backend", hashA, hashB),
			changedLine("backend", hashA, hashB),
		)

		// Both hashes reach the threshold, so both must be reported — not
		// just whichever one happens to have the highest count.
		Expect(utils.DetectXdsChurn(logs)).To(ConsistOf(
			ContainSubstring("hash a1b2c3d4e5f60718"),
			ContainSubstring("hash b1b2c3d4e5f60718"),
		))
	})

	It("ignores malformed lines with no recoverable proxyName", func() {
		logs := joinLines(
			`2026-07-08T00:00:00Z	INFO	xds.reconcile	config has changed	{"versions": ["a1b2c3d4e5f60718"]}`,
			`config has changed with no structured fields at all`,
		)

		Expect(utils.DetectXdsChurn(logs)).To(BeEmpty())
	})

	It("returns a deterministically sorted report for multiple flagged proxies", func() {
		logs := joinLines(
			changedLine("zebra", hashA),
			changedLine("zebra", hashA),
			changedLine("zebra", hashA),
			changedLine("alpha", hashB),
			changedLine("alpha", hashB),
			changedLine("alpha", hashB),
		)

		Expect(utils.DetectXdsChurn(logs)).To(Equal([]string{
			"proxy alpha in mesh default regenerated identical config 3 times (hash b1b2c3d4e5f60718) — non-deterministic xDS",
			"proxy zebra in mesh default regenerated identical config 3 times (hash a1b2c3d4e5f60718) — non-deterministic xDS",
		}))
	})

	It("returns an empty report for empty input", func() {
		Expect(utils.DetectXdsChurn("")).To(BeEmpty())
	})

	It("does not flag a single log line with a duplicate hash in its versions array", func() {
		logs := changedLine("backend", hashA, hashA, hashA)

		Expect(utils.DetectXdsChurn(logs)).To(BeEmpty())
	})

	It("counts a duplicate hash within one line as a single event toward the threshold", func() {
		logs := joinLines(
			changedLine("backend", hashA),
			changedLine("backend", hashA),
			changedLine("backend", hashA, hashA, hashA),
		)

		Expect(utils.DetectXdsChurn(logs)).To(ConsistOf(
			"proxy backend in mesh default regenerated identical config 3 times (hash a1b2c3d4e5f60718) — non-deterministic xDS",
		))
	})

	It("flags the empty-resources hash like any other hash once it reaches the threshold", func() {
		// pkg/xds/server/v3.emptyResourcesVersion() returns a constant hash
		// whenever any resource type transitions from populated to empty, but
		// it gets no special exemption: reappearing 3+ times for one proxy is
		// treated the same as any other hash reaching xdsChurnThreshold.
		logs := joinLines(
			changedLine("backend", "34c96acdcadb1bbb"),
			changedLine("backend", "34c96acdcadb1bbb"),
			changedLine("backend", "34c96acdcadb1bbb"),
		)

		Expect(utils.DetectXdsChurn(logs)).To(ConsistOf(
			ContainSubstring("hash 34c96acdcadb1bbb"),
		))
	})

	It("does not merge same-named proxies across meshes", func() {
		// demo-client exists in two meshes and each regenerates hashA only
		// once. Keyed by name alone the interleaved lines would look like a
		// 3x streak; keyed per (mesh, proxy) neither reaches the threshold.
		logs := joinLines(
			changedLineForDemoClient("mesh-1", hashA),
			changedLineForDemoClient("mesh-2", hashA),
			changedLineForDemoClient("mesh-1", hashB),
			changedLineForDemoClient("mesh-2", hashB),
		)

		Expect(utils.DetectXdsChurn(logs)).To(BeEmpty())
	})

	It("flags the churning mesh's proxy without implicating its same-named twin", func() {
		logs := joinLines(
			changedLineForDemoClient("mesh-1", hashA),
			changedLineForDemoClient("mesh-1", hashA),
			changedLineForDemoClient("mesh-1", hashA),
			changedLineForDemoClient("mesh-2", hashA),
			changedLineForDemoClient("mesh-2", hashB),
		)

		Expect(utils.DetectXdsChurn(logs)).To(ConsistOf(
			"proxy demo-client in mesh mesh-1 regenerated identical config 3 times (hash a1b2c3d4e5f60718) — non-deterministic xDS",
		))
	})
})
